package randoseed

import (
	"net/http"
	"os"
	"path/filepath"
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

func AddRoutes(r *gin.Engine) {
	r.GET("/s/:filehash", getSeed)
	r.GET("/download/:filehash", downloadSeed)

	authGroup := r.Group("/", authRequired())
	authGroup.POST("/uploadseed", uploadSeed)
	authGroup.POST("/vote/:filehash", voteOnSeed)
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

	seed, err := GetByFileHashWithRawSettings(db, filehash)
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

func downloadSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	db := c.Value("database").(*gorm.DB)
	spoilerLogDir := c.Value("filestore").(string)

	_, err := GetByFileHash(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	fileName := filehash + ".json"
	c.FileAttachment(filepath.Join(spoilerLogDir, fileName), fileName)
}

func uploadSeed(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)
	validationError := func(err string) {
		errModel := struct {
			FieldName string
			Error     string
		}{
			FieldName: "",
			Error:     err,
		}
		c.HTML(http.StatusBadRequest, "uploadSeed", []any{
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
		c.Redirect(http.StatusFound, "/s/"+alreadyUploaded.FileHash)
		return
	}

	spoilerlogFile, err := uploadedFile.Open()
	defer spoilerlogFile.Close()
	if err != nil {
		validationError("File did not upload correctly")
		return
	}

	spoilerLog, spoilerLogBytes, jsonErr := GetSpoilerLogFromJsonFile(spoilerlogFile)
	if jsonErr != nil {
		validationError("Spoiler Log JSON could not be read")
		return
	}

	newDbRecord := spoilerLog.CreateDatabaseSeed()

	v := binding.Validator.Engine().(*validator.Validate)
	err = v.Struct(*newDbRecord)
	if err != nil {
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
	c.Redirect(http.StatusFound, redirectDest)
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
