package randoseed

import (
	"net/http"
	"os"
	"strings"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ViewSeedModel struct {
	util.ViewModel
	Seed      *Seed
	AvgRating *AvgSeedRank
	MyRating  *SeedRank
}

// TODO: any way to combine this with ViewUploadCommentModel? Generics maybe?
func (v ViewSeedModel) UserCanEdit() bool {
	return v.User != nil &&
		v.Seed != nil &&
		v.Seed.User != nil &&
		v.User.ID == v.Seed.User.ID
}

func AddRoutes(r *gin.Engine) {
	r.GET("/s/:filehash", getSeed)
	r.GET("/download/:filehash", downloadSeed)
	r.GET("/s/:filehash/uploadercomment", getUploaderComment)

	authGroup := r.Group("/", authRequired())
	authGroup.POST("/uploadseed", uploadSeed)
	authGroup.POST("/vote/:filehash", voteOnSeed)
	authGroup.GET("/s/:filehash/uploadercomment/edit", editUploaderCommentUI)
	authGroup.POST("/s/:filehash/uploadercomment", editUploaderComment)
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authentication.GetCurrentUser(c)

		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func getSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := GetByFileHashWithRelationships(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	avgRating, avgErr := GetAverageRank(db, seed.ID)
	if avgErr != nil {
		c.AbortWithError(http.StatusInternalServerError, avgErr)
		return
	}
	var myRating *SeedRank
	if user != nil {
		myRating, err = GetUserRank(db, seed.ID, user.ID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
	c.HTML(http.StatusOK, "seed.html", ViewSeedModel{
		ViewModel: util.ViewModel{
			User: user,
		},
		Seed:      seed,
		AvgRating: avgRating,
		MyRating:  myRating,
	})
}

type ViewUploaderCommentModel struct {
	Seed  *Seed
	User  *authentication.UserDisplay
	Error string
}

func (v ViewUploaderCommentModel) UserCanEdit() bool {
	return v.User != nil &&
		v.Seed != nil &&
		v.Seed.User != nil &&
		v.User.ID == v.Seed.User.ID
}

func getUploaderComment(c *gin.Context) {
	filehash := c.Param("filehash")
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := GetByFileHashWithRelationships(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.HTML(http.StatusOK, "uploaderComment", ViewUploaderCommentModel{
		Seed: seed,
		User: user,
	})
}

func editUploaderComment(c *gin.Context) {
	filehash := c.Param("filehash")
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := GetByFileHashWithRelationships(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// TODO: validation working?
	if err := c.Bind(seed); err != nil {
		c.HTML(http.StatusOK, "editUploaderComment", ViewUploaderCommentModel{
			Seed:  seed,
			User:  user,
			Error: "Failed validation",
		})
		return
	}

	if err := db.Save(seed).Error; err != nil {
		c.HTML(http.StatusOK, "editUploaderComment", ViewUploaderCommentModel{
			Seed:  seed,
			User:  user,
			Error: "Invalid comment",
		})
		return
	}

	c.HTML(http.StatusOK, "uploaderComment", ViewUploaderCommentModel{
		Seed: seed,
		User: user,
	})
}

func editUploaderCommentUI(c *gin.Context) {
	filehash := c.Param("filehash")
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := GetByFileHashWithRelationships(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.HTML(http.StatusOK, "editUploaderComment", ViewUploaderCommentModel{
		Seed: seed,
		User: user,
	})
}

func downloadSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	db := util.GetDatabase(c)

	_, err := GetByFileHash(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	fileName := filehash + ".json"
	c.FileAttachment(util.GetSpoilerLogDest(c, filehash), fileName)
}

func uploadSeed(c *gin.Context) {
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)
	validationError := func(err string) {
		errModel := struct {
			FieldName string
			Error     string
		}{
			FieldName: "",
			Error:     err,
		}
		c.HTML(http.StatusOK, "uploadSeed", []any{
			errModel,
		})
	}

	// TODO: some kind of CAPTCHA
	form, err := c.MultipartForm()
	if err != nil {
		validationError("No file uploaded")
		return
	}
	formData := form.File["spoilerlog"]
	if len(formData) == 0 {
		validationError("No spoiler log file uploaded")
		return
	}

	uploadedFile := formData[0]
	uploadedFilename := formData[0].Filename

	if alreadyUploaded, _ := GetByFileHash(db, strings.Replace(uploadedFilename, ".json", "", 1)); alreadyUploaded != nil {
		util.HtmxRedirect(c, "/s/"+alreadyUploaded.FileHash)
		return
	}

	spoilerlogFile, err := uploadedFile.Open()
	if err != nil {
		validationError("File did not upload correctly")
		return
	}
	defer spoilerlogFile.Close()

	spoilerLog, spoilerLogBytes, jsonErr := GetSpoilerLogFromJsonFile(spoilerlogFile)
	if jsonErr != nil {
		validationError("Spoiler Log JSON could not be read")
		return
	}

	newDbRecord := spoilerLog.CreateDatabaseSeed(user, "")

	v := binding.Validator.Engine().(*validator.Validate)
	if err = v.Struct(*newDbRecord); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.HTML(http.StatusOK, "uploadSeed", validationErrors)
		return
	}

	// TODO: need to use db.WithContext for proper transaction?
	createResult := db.Create(&newDbRecord)
	if createResult.Error != nil {
		validationError("Seed could not be saved to database")
		return
	}

	writeErr := os.WriteFile(util.GetSpoilerLogDest(c, newDbRecord.FileHash), spoilerLogBytes.Bytes(), 0777)
	if writeErr != nil {
		validationError("Spoiler log could not be uploaded to storage")
		return
	}

	// TODO: use transaction again with gorm
	// if err = createResult.Commit().Error; err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	redirectDest := "/s/" + newDbRecord.FileHash
	util.HtmxRedirect(c, redirectDest)
}

func voteOnSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	db := c.Value("database").(*gorm.DB)
	user := authentication.GetCurrentUser(c)

	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	seed, err := GetByFileHash(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var rank *SeedRank

	if existingRank, getErr := GetUserRank(db, seed.ID, user.ID); err != nil {
		c.AbortWithError(http.StatusInternalServerError, getErr)
		return
	} else if existingRank != nil {
		rank = existingRank
	} else {
		rank = &SeedRank{}
		rank.SeedID = seed.ID
		rank.UserID = user.ID
	}

	// TODO: CAPTCHA

	if err := c.Bind(rank); err != nil {
		return
	}

	if err = db.Save(&rank).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	avgRating, avgErr := GetAverageRank(db, seed.ID)
	if avgErr != nil {
		c.AbortWithError(http.StatusInternalServerError, avgErr)
		return
	}

	c.HTML(http.StatusOK, "seedrank", ViewSeedModel{
		ViewModel: util.ViewModel{
			User: user,
		},
		Seed:      seed,
		AvgRating: avgRating,
		MyRating:  rank,
	})
}
