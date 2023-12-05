package migration

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/scrubsanity"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/tokensanity"
	"github.com/bsinky/sohrando/routes"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Not sure why I couldn't use these from routes package but...I couldn't
func FreshDb(t *testing.T, path ...string) *routes.App {
	t.Helper()

	app := FreshDbWithoutMigrations(t, path...)
	if err := MigrateDB(app.DB); err != nil {
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

	app, err := routes.SetUpDBAndStorage(dbUri)
	if err != nil {
		t.Fatalf("Error opening memory db: %s", err)
	}
	return app
}

func copySpoilerLogToDB(t *testing.T, db *gorm.DB, spoilerFile string, seedID uint) {
	t.Helper()
	srcFile, err := os.Open(path.Join("..", "test", spoilerFile))
	if err != nil {
		t.Fatalf("Error opening %s: %s", spoilerFile, err)
	}
	defer srcFile.Close()

	_, spoilerLogBytes, jsonErr := randoseed.GetSpoilerLogFromJsonFile(srcFile)
	if jsonErr != nil {
		t.Fatalf("Error creating spoiler log from JSON %s", jsonErr)
	}

	newDbSpoilerLogFile := &randoseed.SpoilerLogFile{
		SeedID:         seedID,
		SpoilerLogJSON: datatypes.JSON(spoilerLogBytes.Bytes()),
	}
	if err = db.Create(newDbSpoilerLogFile).Error; err != nil {
		t.Fatalf("Error creating spoiler log file in db %s", err)
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
		Logic       logic.Logic
		Shopsanity  shopsanity.Shopsanity
		Tokensanity tokensanity.Tokensanity
		Scrubsanity scrubsanity.Scrubsanity
	}

	// Create seeds table without all the current columns
	db.Table("seeds").AutoMigrate(&OldSeedDefinition{})
	db.AutoMigrate(&randoseed.SpoilerLogFile{})

	newSeeds := []*OldSeedDefinition{
		{
			Seed:        "1611283300",
			Version:     "KHAN BRAVO (6.1.1)",
			FileHash:    "04-94-01-69-66",
			Logic:       logic.Glitchless,
			Shopsanity:  shopsanity.Random,
			Tokensanity: tokensanity.Off,
			Scrubsanity: scrubsanity.Off,
		},
		{
			Seed:        "4391",
			Version:     "Sulu Bravo (7.1.1)",
			FileHash:    "27-46-32-77-65",
			Logic:       logic.Glitchless,
			Shopsanity:  shopsanity.Random,
			Tokensanity: tokensanity.AllTokens,
			Scrubsanity: scrubsanity.Off,
		},
	}

	for _, seed := range newSeeds {
		if err := db.Table("seeds").Save(seed).Error; err != nil {
			t.Fatalf("Error saving seed before migration %s", err)
		}
		uploadedFileName := seed.FileHash + ".json"
		copySpoilerLogToDB(t, db, uploadedFileName, seed.ID)
	}

	if err := MigrateDB(db); err != nil {
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
			if fieldName == "UploaderComment" {
				// Comment doesn't migrate from the spoiler log file
				continue
			}
			if fieldValue.String() == "" {
				t.Fatalf("Seed %s didn't migrate %s", seed.FileHash, fieldName)
			}
		}
	}
}
