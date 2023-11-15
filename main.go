package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bsinky/sohrando/migration"
	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/routes"
	"github.com/bsinky/sohrando/util"
	"github.com/go-playground/validator/v10"

	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	spoilerLogDir string
	DB            *gorm.DB
}

const sqliteDbFileName = "sqlite.db"

func fileHashIcons(fileHash string) []string {
	hashIconUrls := make([]string, 5)
	for i, hash := range strings.Split(fileHash, "-") {
		hashIconUrls[i] = "/assets/hash/" + hash + ".png"
	}
	return hashIconUrls
}

func preserveLinebreaks(text string) template.HTML {
	return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
}

// TODO: better name and move to different package
type Err1 struct {
	FieldName string
	Error     string
}

func toErrors(errors any) []Err1 {
	errs := make([]Err1, 0)
	switch v := errors.(type) {
	case string:
		if len(v) > 0 {
			return []Err1{
				{
					Error: v,
				},
			}
		}
	case validator.ValidationErrors:
		for _, vErr := range v {
			errs = append(errs, Err1{
				FieldName: vErr.Field(),
				Error:     vErr.Error(),
			})
		}
	}

	return errs
}

func SetUpDBAndStorage(dbURI string, storageDir string) (*App, error) {
	db, err := gorm.Open(sqlite.Open(dbURI), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(storageDir, 0777); err != nil {
		return nil, err
	}

	return &App{
		spoilerLogDir: storageDir,
		DB:            db,
	}, nil
}

func SetupRouter(r *gin.Engine, app *App) {
	r.StaticFS("/assets", http.Dir("assets"))
	r.SetFuncMap(template.FuncMap{
		"fileHashIcons":      fileHashIcons,
		"preserveLinebreaks": preserveLinebreaks,
		"toErrors":           toErrors,
	})
	r.LoadHTMLGlob("templates/*")

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		randoseed.RegisterValidation(v)
	}

	r.Use(util.ConnectDatabase(app.DB))

	// TODO: better secret
	store := gormsessions.NewStore(app.DB, true, []byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.Use(util.ConnectFilestore(app.spoilerLogDir))

	r.GET("/", func(c *gin.Context) {
		seeds, err := randoseed.MostRecent(app.DB, 10)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "index.html", util.ViewData(c, &gin.H{"seeds": seeds}))
	})

	routes.AddSeedRoutes(r)
	routes.AddSearchRoutes(r)
	routes.AddUserRoutes(r)
}

func main() {
	spoilerLogDir := "spoiler_logs"
	app, setupErr := SetUpDBAndStorage(sqliteDbFileName, spoilerLogDir)
	if setupErr != nil {
		log.Fatal(setupErr)
	}

	if err := migration.MigrateDB(app.DB, spoilerLogDir); err != nil {
		log.Fatal(err)
	}

	f, _ := os.Create("oot-rando.log")
	// Write to logfile and stdout
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()
	SetupRouter(r, app)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
