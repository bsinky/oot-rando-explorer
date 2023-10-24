package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ootrandoexplorer/site/randoseed"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var seedsDatabase *randoseed.SQLiteRepository

const spoilerLogDir = "spoiler_logs"

func getSpoilerLogDest(fileHash string) string {
	fileName := fileHash + ".json"
	return filepath.Join(spoilerLogDir, fileName)
}

func getSeed(c *gin.Context) {
	filehash := c.Param("filehash")

	seed, err := seedsDatabase.GetByFileHash(filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.HTML(http.StatusOK, "seed.tmpl", seed)
}

func downloadSeed(c *gin.Context) {
	filehash := c.Param("filehash")

	_, err := seedsDatabase.GetByFileHash(filehash)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	fileName := filehash + ".json"
	c.FileAttachment(filepath.Join(spoilerLogDir, fileName), fileName)
}

func uploadSeed(c *gin.Context) {
	// TODO: some kind of CAPTCHA
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println("Error getting multipart form")
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
	fmt.Printf("multipart Filename is %s", uploadedFilename)

	if alreadyUploaded, _ := seedsDatabase.GetByFileHash(strings.Replace(uploadedFilename, ".json", "", 1)); alreadyUploaded != nil {
		c.Redirect(http.StatusFound, "/s/"+alreadyUploaded.FileHash)
		return
	}

	spoilerlogFile, err := uploadedFile.Open()
	defer spoilerlogFile.Close()
	if err != nil {
		fmt.Println("Couldn't open form file")
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	spoilerLogBytes := bytes.NewBuffer(nil)
	spoilerLogSize, err := io.Copy(spoilerLogBytes, spoilerlogFile)
	if err != nil || spoilerLogSize == 0 {
		fmt.Println("Couldn't read form file")
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	spoilerLog := randoseed.SpoilerLog{}
	jsonErr := json.Unmarshal(spoilerLogBytes.Bytes(), &spoilerLog)
	if jsonErr != nil {
		fmt.Println("Couldn't parse JSON")
		c.AbortWithError(http.StatusBadRequest, jsonErr)
		return
	}

	rawSettings := struct {
		settings any
	}{}
	if sJsonErr := json.Unmarshal(spoilerLogBytes.Bytes(), &rawSettings); sJsonErr != nil {
		fmt.Println("Couldn't parse raw settings json")
		c.AbortWithError(http.StatusBadRequest, sJsonErr)
		return
	}

	// TODO: use a database transaction and only commit it once writing file to disk succeeds too
	// TODO: raw settings JSON isn't working as expected, probably just  write another method to extract it
	rawSettingsJson, rErr := json.Marshal(rawSettings)
	if rErr != nil {
		fmt.Println("Couldn't serialize raw settings json")
		c.AbortWithError(http.StatusBadRequest, rErr)
		return
	}

	newDbRecord := randoseed.MakeDatabaseRecord(spoilerLog, string(rawSettingsJson))
	if _, insertErr := seedsDatabase.Create(newDbRecord); insertErr != nil {
		fmt.Println("Couldn't insert new db record")
		c.AbortWithError(http.StatusInternalServerError, insertErr)
		return
	}

	writeErr := os.WriteFile(getSpoilerLogDest(newDbRecord.FileHash), spoilerLogBytes.Bytes(), 0777)
	if writeErr != nil {
		fmt.Println("Couldn't write spoiler log to disk")
		c.AbortWithError(http.StatusInternalServerError, writeErr)
		return
	}

	redirectDest := "/s/" + newDbRecord.FileHash
	c.Header("HX-Location", redirectDest)
	c.Redirect(http.StatusFound, redirectDest)
}

const sqliteDbFileName = "sqlite.db"

func main() {
	db, err := sql.Open("sqlite3", sqliteDbFileName)
	if err != nil {
		log.Fatal(err)
	}

	seedsDatabase = randoseed.NewSQLiteRepository(db)
	if err := seedsDatabase.Migrate(); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(spoilerLogDir, 0777); err != nil {
		log.Fatal(err)
	}

	f, _ := os.Create("oot-rando.log")
	// Write to logfile and stdout
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()
	r.StaticFS("/assets", http.Dir("assets"))
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		seeds, err := seedsDatabase.MostRecent(10)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "index.tmpl", gin.H{"seeds": seeds})
	})

	r.GET("/s/:filehash", getSeed)
	r.GET("/download/:filehash", downloadSeed)

	r.POST("/uploadseed", uploadSeed)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
