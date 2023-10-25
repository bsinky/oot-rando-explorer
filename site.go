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
	"text/template"

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

	if alreadyUploaded, _ := seedsDatabase.GetByFileHash(strings.Replace(uploadedFilename, ".json", "", 1)); alreadyUploaded != nil {
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

	tx, txErr := seedsDatabase.BeginTx()
	if txErr != nil {
		c.AbortWithError(http.StatusInternalServerError, txErr)
		return
	}
	defer tx.Rollback()

	newDbRecord := randoseed.MakeDatabaseRecord(spoilerLog)
	if _, insertErr := seedsDatabase.Create(newDbRecord, tx); insertErr != nil {
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

const sqliteDbFileName = "sqlite.db"

func fileHashIcons(fileHash string) []string {
	hashIconUrls := make([]string, 5)
	for i, hash := range strings.Split(fileHash, "-") {
		hashIconUrls[i] = "/assets/hash/" + hash + ".png"
	}
	return hashIconUrls
}

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
	r.SetFuncMap(template.FuncMap{
		"fileHashIcons": fileHashIcons,
	})
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
