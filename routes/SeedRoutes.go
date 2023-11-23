package routes

import (
	"net/http"
	"os"
	"strings"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ViewSeedModel struct {
	util.ViewModel
	Seed      *randoseed.Seed
	AvgRating *randoseed.AvgSeedRank
	MyRating  *randoseed.SeedRank
}

// TODO: any way to combine this with ViewUploadCommentModel? Generics maybe?
func (v ViewSeedModel) UserCanEdit() bool {
	return v.User != nil &&
		v.Seed != nil &&
		v.Seed.User != nil &&
		v.User.ID == v.Seed.User.ID
}

func AddSeedRoutes(r *gin.Engine) {
	topSeedRoutes := r.Group("/top")
	topSeedRoutes.GET("/", topSeeds)
	topSeedRoutes.GET("/difficulty/easy", nextTopSeeds(randoseed.EasiestSeeds, "/difficulty/easy", avgSeedDifficulty))
	topSeedRoutes.GET("/difficulty/hard", nextTopSeeds(randoseed.HardestSeeds, "/difficulty/hard", avgSeedDifficulty))
	topSeedRoutes.GET("/fun", nextTopSeeds(randoseed.MostFunSeeds, "/fun", avgSeedFun))
	// No route for most boring seeds...who would want that?

	noAuthRoutes := r.Group("/s/:filehash")
	noAuthRoutes.GET("/", getSeed)
	noAuthRoutes.GET("/download", downloadSeed)
	noAuthRoutes.GET("/uploadercomment", getUploaderComment)

	authGroup := r.Group("/", authRequired())
	authGroup.POST("/uploadseed", uploadSeed)
	authGroupWithFilehash := r.Group("/s/:filehash", authRequired())
	authGroupWithFilehash.POST("/vote", voteOnSeed)
	authGroupWithFilehash.GET("/uploadercomment/edit", editUploaderCommentUI)
	authGroupWithFilehash.POST("/uploadercomment", editUploaderComment)
	authGroupWithFilehash.GET("/confirmdelete", seedPredeleteCheck, confirmDelete)
	authGroupWithFilehash.DELETE("/delete", seedPredeleteCheck, performDelete)
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authentication.GetCurrentUser(c)

		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func fileHashParam(c *gin.Context) string {
	return c.Param("filehash")
}

func getSeed(c *gin.Context) {
	filehash := fileHashParam(c)
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := randoseed.GetByFileHashWithRelationships(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	avgRating, avgErr := randoseed.UpdateAverageRank(db, seed.ID)
	if avgErr != nil {
		c.AbortWithError(http.StatusInternalServerError, avgErr)
		return
	}
	var myRating *randoseed.SeedRank
	if user != nil {
		myRating, err = randoseed.GetUserRank(db, seed.ID, user.ID)
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
	Seed  *randoseed.Seed
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
	filehash := fileHashParam(c)
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := randoseed.GetByFileHashWithRelationships(db, filehash)
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
	filehash := fileHashParam(c)
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := randoseed.GetByFileHashWithRelationships(db, filehash)
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
	filehash := fileHashParam(c)
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	seed, err := randoseed.GetByFileHashWithRelationships(db, filehash)
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
	filehash := fileHashParam(c)
	db := util.GetDatabase(c)

	_, err := randoseed.GetByFileHash(db, filehash)
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
		errModel := util.SimpleValidation{
			Message: err,
		}
		c.HTML(http.StatusOK, "uploadSeed", []util.SimpleValidation{
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

	if alreadyUploaded, _ := randoseed.GetByFileHash(db, strings.Replace(uploadedFilename, ".json", "", 1)); alreadyUploaded != nil {
		util.HtmxRedirect(c, "/s/"+alreadyUploaded.FileHash)
		return
	}

	spoilerlogFile, err := uploadedFile.Open()
	if err != nil {
		validationError("File did not upload correctly")
		return
	}
	defer spoilerlogFile.Close()

	spoilerLog, spoilerLogBytes, jsonErr := randoseed.GetSpoilerLogFromJsonFile(spoilerlogFile)
	if jsonErr != nil {
		validationError("Spoiler Log JSON could not be read")
		return
	} else if spoilerLog == nil {
		validationError("Spoiler Log file could not be read")
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
	filehash := fileHashParam(c)
	db := c.Value("database").(*gorm.DB)
	user := authentication.GetCurrentUser(c)

	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	seed, err := randoseed.GetByFileHash(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var rank *randoseed.SeedRank

	if existingRank, getErr := randoseed.GetUserRank(db, seed.ID, user.ID); err != nil {
		c.AbortWithError(http.StatusInternalServerError, getErr)
		return
	} else if existingRank != nil {
		rank = existingRank
	} else {
		rank = &randoseed.SeedRank{}
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

	avgRating, avgErr := randoseed.UpdateAverageRank(db, seed.ID)
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

func topSeeds(c *gin.Context) {
	c.HTML(http.StatusOK, "top.html", util.ViewData(c, &gin.H{}))
}

type topSeedFunc func(*gorm.DB, int, *float64, *uint) ([]randoseed.AvgSeedRank, error)
type lastValueFunc func(*randoseed.AvgSeedRank) float64

func avgSeedDifficulty(s *randoseed.AvgSeedRank) float64 {
	return s.Difficulty
}

func avgSeedFun(s *randoseed.AvgSeedRank) float64 {
	return s.Fun
}

type lastDisplayed struct {
	ID    *uint    `form:"lastid"`
	Value *float64 `form:"lastvalue"`
}

const seedsPerBatch = 10

func nextTopSeeds(topSeeds topSeedFunc, loadMoreAction string, getLastValue lastValueFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := util.GetDatabase(c)
		lastDisplayed := lastDisplayed{}
		if err := c.ShouldBindQuery(&lastDisplayed); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		avgSeedRanks, err := topSeeds(db, seedsPerBatch, lastDisplayed.Value, lastDisplayed.ID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if len(avgSeedRanks) > 0 {
			lastShown := &avgSeedRanks[len(avgSeedRanks)-1]
			val := getLastValue(lastShown)
			lastDisplayed.Value = &val
			lastDisplayed.ID = &lastShown.ID
		} else {
			lastDisplayed.Value = nil
			lastDisplayed.ID = nil
		}

		c.HTML(http.StatusOK, "topseeds", gin.H{
			"AvgSeedRanks":   avgSeedRanks,
			"LoadMoreAction": "/top" + loadMoreAction,
			"LastValue":      lastDisplayed.Value,
			"LastID":         lastDisplayed.ID,
		})
	}
}

func seedPredeleteCheck(c *gin.Context) {
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)
	filehash := fileHashParam(c)

	if !user.CanDeleteSeeds() {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	seed, err := randoseed.GetByFileHash(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if seed == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Set("deleteseed", seed)
}

func confirmDelete(c *gin.Context) {
	seed := c.Value("deleteseed").(*randoseed.Seed)

	c.HTML(http.StatusOK, "seedconfirmdelete", gin.H{
		"Seed": seed,
	})
}

func performDelete(c *gin.Context) {
	db := util.GetDatabase(c)
	seed := c.Value("deleteseed").(*randoseed.Seed)

	if seed == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if err := db.Delete(seed).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	util.HtmxRedirect(c, "/")
}
