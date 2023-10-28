package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"ootrandoexplorer/site/randoseed"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const spoilerLogDir = "spoiler_logs"

func getSpoilerLogDest(fileHash string) string {
	fileName := fileHash + ".json"
	return filepath.Join(spoilerLogDir, fileName)
}

// Middleware to connect the database for each request that uses this
// middleware.
func connectDatabase(db *randoseed.SQLiteRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("database", db)
	}
}

func authenticateUser(db *randoseed.SQLiteRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user *randoseed.User
		auth, err := c.Cookie("auth")
		if err != nil {
			// cookie not found, generate user
			if newUserID, uuidErr := uuid.NewRandom(); uuidErr != nil {
				c.AbortWithError(http.StatusInternalServerError, uuidErr)
				return
			} else {
				auth = newUserID.String()
				c.SetCookie("auth", auth, 60*60*24*30, "/", "localhost", false, false)
			}
		} else {
			// Got an auth cookie from the client, check if it's valid
			user, err = db.GetUser(auth)
			if err != nil && !errors.Is(err, randoseed.ErrNotExists) {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
		}

		if user == nil {
			user = &randoseed.User{
				Username: auth,
			}
			if err = db.CreateUser(user, nil); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		}

		c.Set("user", user)
	}
}

type ViewSeedModel struct {
	Seed      *randoseed.DBSeed
	AvgRating *randoseed.AvgSeedRank
	MyRating  *randoseed.SeedRank
}

func getSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	user := c.Value("user").(*randoseed.User)
	db := c.Value("database").(*randoseed.SQLiteRepository)

	seed, err := db.GetByFileHash(filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	avgRating, avgErr := db.GetAverageRank(filehash)
	if avgErr != nil {
		c.AbortWithError(http.StatusInternalServerError, avgErr)
		return
	}
	myRating, myRatingErr := db.GetUserRank(filehash, user.ID)
	if myRatingErr != nil {
		c.AbortWithError(http.StatusInternalServerError, myRatingErr)
		return
	}

	c.HTML(http.StatusOK, "seed.html", ViewSeedModel{
		Seed:      seed,
		AvgRating: avgRating,
		MyRating:  myRating,
	})
}

func downloadSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	db := c.Value("database").(*randoseed.SQLiteRepository)

	_, err := db.GetByFileHash(filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	fileName := filehash + ".json"
	c.FileAttachment(filepath.Join(spoilerLogDir, fileName), fileName)
}

func uploadSeed(c *gin.Context) {
	db := c.Value("database").(*randoseed.SQLiteRepository)

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

	if alreadyUploaded, _ := db.GetByFileHash(strings.Replace(uploadedFilename, ".json", "", 1)); alreadyUploaded != nil {
		c.Redirect(http.StatusFound, "/s/"+alreadyUploaded.FileHash)
		return
	}

	spoilerlogFile, err := uploadedFile.Open()
	defer spoilerlogFile.Close()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	spoilerLogBytes := bytes.NewBuffer(nil)
	spoilerLogSize, err := io.Copy(spoilerLogBytes, spoilerlogFile)
	if err != nil || spoilerLogSize == 0 {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	spoilerLog := randoseed.SpoilerLog{}
	jsonErr := json.Unmarshal(spoilerLogBytes.Bytes(), &spoilerLog)
	if jsonErr != nil {
		c.AbortWithError(http.StatusBadRequest, jsonErr)
		return
	}

	tx, txErr := db.BeginTx()
	if txErr != nil {
		c.AbortWithError(http.StatusInternalServerError, txErr)
		return
	}
	defer tx.Rollback()

	newDbRecord := randoseed.MakeDatabaseRecord(spoilerLog)
	if _, insertErr := db.CreateSeed(newDbRecord, tx); insertErr != nil {
		c.AbortWithError(http.StatusInternalServerError, insertErr)
		return
	}

	writeErr := os.WriteFile(getSpoilerLogDest(newDbRecord.FileHash), spoilerLogBytes.Bytes(), 0777)
	if writeErr != nil {
		c.AbortWithError(http.StatusInternalServerError, writeErr)
		return
	}

	if err = tx.Commit(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	redirectDest := "/s/" + newDbRecord.FileHash
	c.Redirect(http.StatusFound, redirectDest)
}

func voteOnSeed(c *gin.Context) {
	filehash := c.Param("filehash")
	user := c.Value("user").(*randoseed.User)
	db := c.Value("database").(*randoseed.SQLiteRepository)

	seed, err := db.GetByFileHash(filehash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var rank *randoseed.SeedRank

	if existingRank, getErr := db.GetUserRank(filehash, user.ID); err != nil {
		c.AbortWithError(http.StatusInternalServerError, getErr)
		return
	} else if existingRank != nil {
		rank = existingRank
	} else {
		rank = &randoseed.SeedRank{}
		rank.DBSeedID = seed.Id
		rank.UserID = user.ID
	}

	// TODO: CAPTCHA

	if err := c.Bind(rank); err != nil {
		return
	}

	if rank == nil || rank.ID == 0 {
		// User has not yet voted
		if _, err := db.CreateRank(*rank, nil); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else {
		// User has already voted, update existing vote
		if err := db.UpdateRank(rank); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	avgRating, avgErr := db.GetAverageRank(filehash)
	if avgErr != nil {
		c.AbortWithError(http.StatusInternalServerError, avgErr)
		return
	}

	c.HTML(http.StatusOK, "seedrank", ViewSeedModel{
		Seed:      seed,
		AvgRating: avgRating,
		MyRating:  rank,
	})
}

const sqliteDbFileName = "sqlite.db"

func fileHashIcons(fileHash string) []string {
	hashIconUrls := make([]string, 5)
	for i, hash := range strings.Split(fileHash, "-") {
		hashIconUrls[i] = "/assets/hash/" + hash + ".png"
	}
	return hashIconUrls
}

func SetUpDBAndStorage(dbURI string, storageDir string) (*randoseed.SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbURI)
	if err != nil {
		return nil, err
	}

	repo := randoseed.NewSQLiteRepository(db)
	if err := repo.Migrate(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(storageDir, 0777); err != nil {
		return nil, err
	}

	return repo, nil
}

func SetupRouter(r *gin.Engine, db *randoseed.SQLiteRepository) {
	r.StaticFS("/assets", http.Dir("assets"))
	r.SetFuncMap(template.FuncMap{
		"fileHashIcons": fileHashIcons,
	})
	r.LoadHTMLGlob("templates/*")
	r.Use(connectDatabase(db))
	r.Use(authenticateUser(db))

	r.GET("/", func(c *gin.Context) {
		seeds, err := db.MostRecent(10)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"seeds": seeds})
	})

	r.GET("/s/:filehash", getSeed)
	r.GET("/download/:filehash", downloadSeed)

	r.POST("/uploadseed", uploadSeed)
	r.POST("/vote/:filehash", voteOnSeed)
}

func main() {
	db, setupErr := SetUpDBAndStorage(sqliteDbFileName, spoilerLogDir)
	if setupErr != nil {
		log.Fatal(setupErr)
	}

	f, _ := os.Create("oot-rando.log")
	// Write to logfile and stdout
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()
	SetupRouter(r, db)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
