package migration

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/bsinky/sohrando/randoseed"

	"gorm.io/gorm"
)

type MigrateSeed struct {
	columnsBeforeMigration int
}

func numberOfSeedColumns(db *gorm.DB) (int, error) {
	if !db.Migrator().HasTable(&randoseed.Seed{}) {
		return 0, nil
	}
	var seedColumns []gorm.ColumnType
	seedColumns, err := db.Migrator().ColumnTypes(&randoseed.Seed{})
	if err != nil {
		return 0, err
	}
	return len(seedColumns), nil
}

func (m *MigrateSeed) BeforeAutoMigrate(db *gorm.DB) error {
	if columnsBeforeMigration, err := numberOfSeedColumns(db); err != nil {
		return err
	} else {
		m.columnsBeforeMigration = columnsBeforeMigration
	}
	return nil
}

func (m *MigrateSeed) Migrate(db *gorm.DB, storageDir string) error {
	if m.columnsBeforeMigration == 0 {
		// Table didn't exist before, nothing to migrate
		return nil
	}

	var (
		columnsAfterMigration int
		err                   error
	)
	if columnsAfterMigration, err = numberOfSeedColumns(db); err != nil {
		return err
	}

	if m.columnsBeforeMigration != columnsAfterMigration {
		// Column definitions have changed
		return UpdateAllSeedsFromStoredSpoilerLogs(db, storageDir)
	}

	return nil
}

// Scan stored SpoilerLogs and update all Seeds in the database
func UpdateAllSeedsFromStoredSpoilerLogs(db *gorm.DB, storageDir string) error {
	if dirEntries, err := os.ReadDir(storageDir); err != nil {
		return err
	} else {
		for _, entry := range dirEntries {
			fileName := entry.Name()
			fileHash, isJsonFile := strings.CutSuffix(fileName, ".json")
			if !isJsonFile {
				continue
			}

			seed, seedErr := randoseed.GetByFileHash(db, fileHash)
			if seedErr != nil {
				return seedErr
			}

			jsonFile, fileErr := os.Open(path.Join(storageDir, fileName))
			if fileErr != nil {
				return fileErr
			}
			defer jsonFile.Close()

			spoilerLog, _, spoilerErr := randoseed.GetSpoilerLogFromJsonFile(jsonFile)
			if spoilerErr != nil {
				return spoilerErr
			} else if spoilerLog == nil {
				return errors.New(fmt.Sprintf("Unable to get SpoilerLog from %s", fileName))
			}

			spoilerLog.UpdateDatabaseSeed(seed)

			if err := db.Save(seed).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
