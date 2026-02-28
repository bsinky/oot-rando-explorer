package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/migration"
	"github.com/bsinky/sohrando/routes"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

const DBURIENV string = "OOTRANDODBURI"

func main() {
	dbURI := os.Getenv(DBURIENV)
	if dbURI == "" {
		log.Default().Println(DBURIENV + " not set, using default URI")
		dbURI = "sqlite.db"
	}

	app, setupErr := routes.SetUpDBAndStorage(dbURI)
	if setupErr != nil {
		log.Fatal(setupErr)
	}

	if err := migration.MigrateDB(app.DB); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		handleCommand(app.DB, os.Args[1:])
		return
	}

	// Write to stdout
	gin.DefaultWriter = os.Stdout

	r := gin.Default()
	routes.SetupRouter(r, app)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func handleCommand(db *gorm.DB, args []string) {
	switch args[0] {
	case "make-admin":
		if len(args) != 2 {
			usage("make-admin <username>")
		} else if err := authentication.SetAdmin(db, args[1], true); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("user %s is now admin\n", args[1])
		}
	case "make-moderator":
		if len(args) != 2 {
			usage("make-moderator <username>")
		} else if err := authentication.SetModerator(db, args[1], true); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("user %s is now moderator\n", args[1])
		}
	case "reset-password":
		if len(args) != 3 {
			usage("reset-password <username> <new_password>")
		} else if err := authentication.ResetPassword(db, args[1], args[2]); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("password for %s updated\n", args[1])
		}
	default:
		usage("")
	}
	os.Exit(0)
}

func usage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintln(os.Stderr, "usage: sohrando <command> [args]")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  make-admin <username>")
	fmt.Fprintln(os.Stderr, "  make-moderator <username>")
	fmt.Fprintln(os.Stderr, "  reset-password <username> <new_password>")
	os.Exit(1)
}
