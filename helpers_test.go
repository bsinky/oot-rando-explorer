package main

import (
	"os"
	"path"
	"testing"

	"github.com/bsinky/sohrando/migration"
)

var SpoilerSeedsDir string = "test/spoiler_logs/"

func DeleteUploadedTestSeed(t *testing.T, fileName string) {
	t.Helper()

	if err := os.Remove(path.Join(SpoilerSeedsDir, fileName)); err != nil {
		t.Fatal(err)
	}
}

func FreshDb(t *testing.T, path ...string) *App {
	t.Helper()

	app := FreshDbWithoutMigrations(t, path...)
	if err := migration.MigrateDB(app.DB, SpoilerSeedsDir); err != nil {
		t.Fatalf("Failed to migrate db %s", err)
	}
	return app
}

func FreshDbWithoutMigrations(t *testing.T, path ...string) *App {
	t.Helper()

	var dbUri string

	// Note: path can be specified in an individual test for debugging
	// purposes -- so the db file can be inspected after the test runs.
	// Normally it should be left off so that a truly fresh memory db is
	// used every time.
	if len(path) == 0 {
		dbUri = ":memory:"
	} else {
		dbUri = path[0]
	}

	app, err := SetUpDBAndStorage(dbUri, SpoilerSeedsDir)
	if err != nil {
		t.Fatalf("Error opening memory db: %s", err)
	}
	return app
}
