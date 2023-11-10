package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/migration"
	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/search"
	"github.com/go-playground/validator/v10"

	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getSpoilerLogDest(c *gin.Context, fileHash string) string {
	spoilerLogDir := c.Value("filestore").(string)
	fileName := fileHash + ".json"
	return filepath.Join(spoilerLogDir, fileName)
}

// Middleware to connect the database for each request that uses this
// middleware.
func connectDatabase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("database", db)
	}
}

func connectFilestore(spoilerLogDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("filestore", spoilerLogDir)
	}
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getCurrentUser(c)

		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// func authenticateUser(db *gorm.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		var user *authentication.User
// 		auth, err := c.Cookie("auth")
// 		if err != nil {
// 			// cookie not found, generate user
// 			if newUserID, uuidErr := uuid.NewRandom(); uuidErr != nil {
// 				c.AbortWithError(http.StatusInternalServerError, uuidErr)
// 				return
// 			} else {
// 				auth = newUserID.String()
// 				// TODO: use gin-sessions
// 				c.SetCookie("auth", auth, 60*60*24*30, "/", "localhost", false, false)
// 			}
// 		} else {
// 			// Got an auth cookie from the client, check if it's valid
// 			user, err = authentication.GetUser(db, auth)
// 			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 				c.AbortWithError(http.StatusInternalServerError, err)
// 				return
// 			}
// 		}

// 		if user == nil {
// 			user = &authentication.User{
// 				Username: auth,
// 			}
// 			if err = db.Save(user).Error; err != nil {
// 				c.AbortWithError(http.StatusInternalServerError, err)
// 				return
// 			}
// 		}

// 		c.Set("user", user)
// 	}
//

type SearchModel struct {
	ViewModel
	Filters map[string]*search.SearchFilter
}

func searchPage(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)
	user := getCurrentUser(c)

	allFilters, err := search.AllFilters(db)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "search.html", SearchModel{
		ViewModel: ViewModel{
			User: user,
		},
		Filters: allFilters,
	})
}

type SearchResultModel struct {
	Result *search.Result
}

func runSearch(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)

	allFilters, err := search.AllFilters(db)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	reqFields := make(map[string]string)
	if err := c.BindQuery(reqFields); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var reqFilters []search.SearchFilterValue

	for k, v := range reqFields {
		// Remove invalid field names or blank values
		if fieldFilter, ok := allFilters[k]; !ok || v == "" {
			continue
		} else if !fieldFilter.IsValidOption(v) {
			continue
		}

		reqFilters = append(reqFilters, search.SearchFilterValue{
			FieldName: k,
			Value:     v,
		})
	}

	result, err := search.RunSearch(db, reqFilters)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "searchresult", SearchResultModel{
		Result: result,
	})
}

type ViewSeedModel struct {
	ViewModel
	Seed      *randoseed.Seed
	AvgRating *randoseed.AvgSeedRank
	MyRating  *randoseed.SeedRank
}

func getSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	db := c.Value("database").(*gorm.DB)
	user := getCurrentUser(c)

	seed, err := randoseed.GetByFileHashWithRawSettings(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	avgRating, avgErr := randoseed.GetAverageRank(db, seed.ID)
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
		ViewModel: ViewModel{
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

	_, err := randoseed.GetByFileHash(db, filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	fileName := filehash + ".json"
	c.FileAttachment(filepath.Join(spoilerLogDir, fileName), fileName)
}

func uploadSeed(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)

	// TODO: some kind of CAPTCHA
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	formData := form.File["spoilerlog"]
	if len(formData) == 0 {
		fmt.Println("multipart file has 0 parts")
		return
	}

	uploadedFile := formData[0]
	uploadedFilename := formData[0].Filename

	if alreadyUploaded, _ := randoseed.GetByFileHash(db, strings.Replace(uploadedFilename, ".json", "", 1)); alreadyUploaded != nil {
		c.Redirect(http.StatusFound, "/s/"+alreadyUploaded.FileHash)
		return
	}

	spoilerlogFile, err := uploadedFile.Open()
	defer spoilerlogFile.Close()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	spoilerLog, spoilerLogBytes, jsonErr := randoseed.GetSpoilerLogFromJsonFile(spoilerlogFile)
	if jsonErr != nil {
		c.AbortWithError(http.StatusBadRequest, jsonErr)
		return
	}

	newDbRecord := spoilerLog.CreateDatabaseSeed()

	v := binding.Validator.Engine().(*validator.Validate)
	err = v.Struct(*newDbRecord)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.AbortWithError(http.StatusBadRequest, validationErrors)
		return
	}

	// TODO: need to use db.WithContext for proper transaction?
	createResult := db.Create(&newDbRecord)
	if createResult.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, createResult.Error)
		return
	}

	writeErr := os.WriteFile(getSpoilerLogDest(c, newDbRecord.FileHash), spoilerLogBytes.Bytes(), 0777)
	if writeErr != nil {
		c.AbortWithError(http.StatusInternalServerError, writeErr)
		return
	}

	// TODO: use transaction again with gorm
	// if err = createResult.Commit().Error; err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	redirectDest := "/s/" + newDbRecord.FileHash
	c.Redirect(http.StatusSeeOther, redirectDest)
}

func voteOnSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	db := c.Value("database").(*gorm.DB)
	user := getCurrentUser(c)

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

	avgRating, avgErr := randoseed.GetAverageRank(db, seed.ID)
	if avgErr != nil {
		c.AbortWithError(http.StatusInternalServerError, avgErr)
		return
	}

	c.HTML(http.StatusOK, "seedrank", ViewSeedModel{
		ViewModel: ViewModel{
			User: user,
		},
		Seed:      seed,
		AvgRating: avgRating,
		MyRating:  rank,
	})
}

func loginPage(c *gin.Context) {
	if getCurrentUser(c) != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "login.html", nil)
}

func logoutAction(c *gin.Context) {
	if err := logoutCurrentUser(c); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func getCurrentUser(c *gin.Context) *authentication.UserDisplay {
	session := sessions.Default(c)
	maybeID := session.Get("User.ID")
	maybeUsername := session.Get("User.Username")
	maybeAvatar := session.Get("User.Avatar")
	if maybeUsername == nil || maybeAvatar == nil || maybeID == nil {
		return nil
	}
	return &authentication.UserDisplay{
		ID:       maybeID.(uint),
		Username: maybeUsername.(string),
		Avatar:   maybeAvatar.(string),
	}
}

func setCurrentUser(c *gin.Context, user *authentication.User) error {
	session := sessions.Default(c)
	session.Set("User.ID", user.ID)
	session.Set("User.Username", user.Username)
	session.Set("User.Avatar", user.Avatar)
	return session.Save()
}

func logoutCurrentUser(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}

func loginGetAuthToken(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)

	userForm := &authentication.UserForm{}
	if err := c.Bind(userForm); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := authentication.GetUser(db, userForm.Username)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ok, err := user.PasswordMatches(userForm.Password)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	} else if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := setCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Login successful, redirect back to main page
	c.Redirect(http.StatusSeeOther, "/")
}

func signupPage(c *gin.Context) {
	if getCurrentUser(c) != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "signup.html", nil)
}

func signupCreateUser(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)

	userForm := &authentication.UserForm{}
	if err := c.Bind(userForm); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := authentication.CreateUser(db, userForm)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	} else if user == nil || user.ID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := setCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Registration successful, redirect back to main page
	c.Redirect(http.StatusSeeOther, "/")
}

const sqliteDbFileName = "sqlite.db"

type ViewModel struct {
	User *authentication.UserDisplay
}

func viewData(c *gin.Context, data *gin.H) *gin.H {
	(*data)["User"] = getCurrentUser(c)
	return data
}

func fileHashIcons(fileHash string) []string {
	hashIconUrls := make([]string, 5)
	for i, hash := range strings.Split(fileHash, "-") {
		hashIconUrls[i] = "/assets/hash/" + hash + ".png"
	}
	return hashIconUrls
}

func SetUpDBAndStorage(dbURI string, storageDir string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbURI), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(storageDir, 0777); err != nil {
		return nil, err
	}

	return db, nil
}

func SetupRouter(r *gin.Engine, db *gorm.DB, spoilerLogDir string) {
	r.StaticFS("/assets", http.Dir("assets"))
	r.SetFuncMap(template.FuncMap{
		"fileHashIcons": fileHashIcons,
	})
	r.LoadHTMLGlob("templates/*")

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		randoseed.RegisterValidation(v)
	}

	r.Use(connectDatabase(db))

	// TODO: better secret
	store := gormsessions.NewStore(db, true, []byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.Use(connectFilestore(spoilerLogDir))

	r.GET("/", func(c *gin.Context) {
		seeds, err := randoseed.MostRecent(db, 10)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "index.html", viewData(c, &gin.H{"seeds": seeds}))
	})

	r.GET("/search", searchPage)
	r.GET("/search/run", runSearch)
	r.GET("/s/:filehash", getSeed)
	r.GET("/download/:filehash", downloadSeed)
	r.GET("/login", loginPage)
	r.GET("/logout", logoutAction)

	r.POST("/login/auth", loginGetAuthToken)
	r.POST("/signup/register", signupCreateUser)

	authGroup := r.Group("/", authRequired())
	authGroup.POST("/uploadseed", uploadSeed)
	authGroup.POST("/vote/:filehash", voteOnSeed)
}

func main() {
	spoilerLogDir := "spoiler_logs"
	db, setupErr := SetUpDBAndStorage(sqliteDbFileName, spoilerLogDir)
	if setupErr != nil {
		log.Fatal(setupErr)
	}

	if err := migration.MigrateDB(db, spoilerLogDir); err != nil {
		log.Fatal(err)
	}

	f, _ := os.Create("oot-rando.log")
	// Write to logfile and stdout
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()
	SetupRouter(r, db, spoilerLogDir)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
