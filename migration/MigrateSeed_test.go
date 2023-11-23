package migration

import (
	"io"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/routes"
	"gorm.io/gorm"
)

var SpoilerSeedsDir string

// Not sure why I couldn't use these from routes package but...I couldn't
func FreshDb(t *testing.T, path ...string) *routes.App {
	t.Helper()
	SpoilerSeedsDir = t.TempDir()

	app := FreshDbWithoutMigrations(t, path...)
	if err := MigrateDB(app.DB, SpoilerSeedsDir); err != nil {
		t.Fatalf("Failed to migrate db %s", err)
	}

	return app
}

func FreshDbWithoutMigrations(t *testing.T, path ...string) *routes.App {
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

	app, err := routes.SetUpDBAndStorage(dbUri, SpoilerSeedsDir)
	if err != nil {
		t.Fatalf("Error opening memory db: %s", err)
	}
	return app
}

func copySpoilerLogToTestDir(t *testing.T, spoilerFile string) {
	t.Helper()
	srcFile, err := os.Open(path.Join("test", spoilerFile))
	if err != nil {
		t.Fatalf("Error opening %s: %s", spoilerFile, err)
	}
	defer srcFile.Close()

	dstFileName := path.Join(SpoilerSeedsDir, spoilerFile)
	dstFile, err := os.Create(dstFileName)
	if err != nil {
		t.Fatalf("Unable to create file %s: %s", dstFileName, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		t.Fatalf("Error copying to %s: %s", dstFile.Name(), err)
	}
}

func TestAddingSettingsColumnsMigratesProperly(t *testing.T) {
	app := FreshDbWithoutMigrations(t)
	db := app.DB

	type OldSeedDefinition struct {
		gorm.Model
		Seed        string
		Version     string
		FileHash    string
		Logic       string
		Shopsanity  string
		Tokensanity string
		Scrubsanity string
	}

	// Create seeds table without all the current columns
	db.Table("seeds").AutoMigrate(&OldSeedDefinition{})

	newSeeds := []*OldSeedDefinition{
		{
			Seed:        "1611283300",
			Version:     "KHAN BRAVO (6.1.1)",
			FileHash:    "04-94-01-69-66",
			Logic:       "Glitchless",
			Shopsanity:  "Random",
			Tokensanity: "Off",
			Scrubsanity: "Off",
		},
		{
			Seed:        "4391",
			Version:     "Sulu Bravo (7.1.1)",
			FileHash:    "27-46-32-77-65",
			Logic:       "Glitchless",
			Shopsanity:  "Random",
			Tokensanity: "All Tokens",
			Scrubsanity: "Off",
		},
	}

	for _, seed := range newSeeds {
		uploadedFileName := seed.FileHash + ".json"
		copySpoilerLogToTestDir(t, uploadedFileName)
		if err := db.Table("seeds").Save(seed).Error; err != nil {
			t.Fatalf("Error saving seed before migration %s", err)
		}
	}

	if err := MigrateDB(db, SpoilerSeedsDir); err != nil {
		t.Fatal(err)
	}

	seedType := reflect.TypeOf(randoseed.Seed{})
	oldFields := reflect.VisibleFields(reflect.TypeOf(OldSeedDefinition{}))
	currentFields := reflect.VisibleFields(seedType)
	var fieldsToCheck []string

	for _, currentField := range currentFields {
		found := false
		for _, oldField := range oldFields {
			if oldField.Name == currentField.Name {
				found = true
				break
			}
		}

		if !found {
			// Field is in current definition but not old one
			fieldsToCheck = append(fieldsToCheck, currentField.Name)
		}
	}

	for _, seed := range newSeeds {
		seedFromDB, err := randoseed.GetByFileHash(db, seed.FileHash)
		if err != nil {
			t.Fatal(err)
		} else if seedFromDB == nil {
			t.Fatalf("Unable to retrieve seed %s from db", seed.FileHash)
		}

		rSeed := reflect.Indirect(reflect.ValueOf(seedFromDB))
		for _, fieldName := range fieldsToCheck {
			fieldValue := rSeed.FieldByName(fieldName)
			if fieldValue.String() == "" {
				t.Fatalf("Seed %s didn't migrate %s", seed.FileHash, fieldName)
			}
		}
	}
}
