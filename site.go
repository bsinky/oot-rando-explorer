package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// JSON-parse types
type SpoilerLog struct {
	Seed     string        `json:"_seed"`
	Version  string        `json:"_version"`
	FileHash []int         `json:"file_hash"`
	Settings RandoSettings `json:"settings"`
}

type RandoSettings struct {
	Logic       string `json:"Logic Options:Logic"`
	Shopsanity  string `json:"Shuffle Settings:Shopsanity"`
	Tokensanity string `json:"Shuffle Settings:Tokensanity"`
	Scrubsanity string `json:"Shuffle Settings:Scrub Shuffle"`
}

// database types
type DBSeed struct {
	Seed        string
	Version     string
	FileHash    string
	Logic       string
	Shopsanity  string
	Tokensanity string
	Scrubsanity string
	RawSettings string
}

func (s SpoilerLog) FileHashString() string {
	hashString := strings.Builder{}
	for i := 0; i < len(s.FileHash); i++ {
		if s.FileHash[i] < 10 {
			hashString.WriteString("0")
		}
		hashString.WriteString(strconv.Itoa((s.FileHash[i])))
		hashString.WriteString("-")
	}
	ret := hashString.String()
	return ret[:len(ret)-1] // remove trailing "-"
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
	// TODO: check spoiler log file name, see if an existing uploaded seed matches it and redirect to that page if so

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

	// fmt.Printf("Attempting to parse json from %s", spoilerLogBytes[0:30])
	spoilerLog := SpoilerLog{}
	jsonErr := json.Unmarshal(spoilerLogBytes.Bytes(), &spoilerLog)
	if jsonErr != nil {
		fmt.Println("Couldn't parse JSON")
		c.AbortWithError(http.StatusBadRequest, jsonErr)
		return
	}

	// TODO: create database records
	// TODO: write uploaded file to file system
}

func main() {
	r := gin.Default()
	r.StaticFS("/assets", http.Dir("assets"))
	r.LoadHTMLFiles("index.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/uploadseed", uploadSeed)

	r.GET("/download/:seedhash", func(c *gin.Context) {
		// c.FileAttachment()
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
