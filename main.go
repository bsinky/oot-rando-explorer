package main

import (
	"io"
	"log"
	"os"

	"github.com/bsinky/sohrando/migration"
	"github.com/bsinky/sohrando/routes"

	"github.com/gin-gonic/gin"
)

const sqliteDbFileName = "sqlite.db"

func main() {
	app, setupErr := routes.SetUpDBAndStorage(sqliteDbFileName)
	if setupErr != nil {
		log.Fatal(setupErr)
	}

	if err := migration.MigrateDB(app.DB); err != nil {
		log.Fatal(err)
	}

	f, _ := os.Create("oot-rando.log")
	// Write to logfile and stdout
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()
	routes.SetupRouter(r, app)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
