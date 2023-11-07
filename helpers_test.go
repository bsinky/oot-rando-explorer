package main

import (
	"os"
	"path"
	"testing"

	"github.com/bsinky/sohrando/migration"
	"gorm.io/gorm"
)

var SpoilerSeedsDir string = "test/spoiler_logs/"

func DeleteUploadedTestSeed(t *testing.T, fileName string) {
	t.Helper()

	if err := os.Remove(path.Join(SpoilerSeedsDir, fileName)); err != nil {
		t.Fatal(err)
	}
}

func FreshDb(t *testing.T, path ...string) *gorm.DB {
	t.Helper()

	db := FreshDbWithoutMigrations(t, path...)
	if err := migration.MigrateDB(db, SpoilerSeedsDir); err != nil {
		t.Fatalf("Failed to migrate db %s", err)
	}
	return db
}

func FreshDbWithoutMigrations(t *testing.T, path ...string) *gorm.DB {
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

	db, err := SetUpDBAndStorage(dbUri, SpoilerSeedsDir)
	if err != nil {
		t.Fatalf("Error opening memory db: %s", err)
	}
	return db
}
