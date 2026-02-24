package routes

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/util"
	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"github.com/inhies/go-bytesize"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	DB *gorm.DB
}

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

func getHtmlStyles() string {
	return "dark"
}

func getBodyStyles() string {
	return "bg-zinc-950 text-zinc-100 min-h-screen font-sans selection:bg-orange-500/30"
}

func templateDicts(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func SetUpDBAndStorage(dbURI string) (*App, error) {
	db, err := gorm.Open(getDBProvider(dbURI), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &App{
		DB: db,
	}, nil
}

func getDBProvider(dbURI string) gorm.Dialector {
	const envKey string = "OOTRANDODB"
	switch strings.ToLower(os.Getenv(envKey)) {
	case "sqlite":
		return sqlite.Open(dbURI)
	case "postgres":
		return postgres.Open(dbURI)
	default:
		log.Default().Println(envKey + " not set or unknown, falling back to SQLite")
		return sqlite.Open(dbURI)
	}
}

func getTemplatesDir() string {
	var templatesDir string
	if templatesDirFromEnv := os.Getenv("OOTRANDOTEMPLATESDIR"); templatesDirFromEnv != "" {
		templatesDir = templatesDirFromEnv
	} else {
		templatesDir = "./templates"
	}
	return path.Join(templatesDir, "*")
}

func getSessionSecret() string {
	envKey := "OOTRANDOSESSIONSECRET"
	secret := os.Getenv(envKey)
	if secret == "" {
		log.Default().Println(envKey + " not set, using default session secret")
		return "secret"
	}
	return secret
}

func SetupRouter(r *gin.Engine, app *App) {
	randoseed.InitVersionCache(app.DB)
	r.StaticFS("/assets", http.Dir("assets"))
	r.SetFuncMap(template.FuncMap{
		"fileHashIcons":      fileHashIcons,
		"getHtmlStyles":      getHtmlStyles,
		"getBodyStyles":      getBodyStyles,
		"dict":               templateDicts,
		"preserveLinebreaks": preserveLinebreaks,
		"toErrors":           util.ToErrors,
	})

	r.LoadHTMLGlob(getTemplatesDir())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		randoseed.RegisterValidation(v)
	}

	// Limit requests to 5MB, uploaded seeds should all be well below that
	r.Use(limits.RequestSizeLimiter(int64(bytesize.MB * 5)))
	r.Use(util.ConnectDatabase(app.DB))

	store := gormsessions.NewStore(app.DB, true, []byte(getSessionSecret()))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/", func(c *gin.Context) {
		seeds, err := randoseed.MostRecent(app.DB, 10)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "index.html", util.ViewData(c, &gin.H{"seeds": seeds}))
	})

	AddSeedRoutes(r)
	AddSearchRoutes(r)
	AddUserRoutes(r)
}
