package migration

import (
	"fmt"
	"log"

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

func (m *MigrateSeed) Migrate(db *gorm.DB) error {
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
		return updateAllSeedsFromStoredSpoilerLogs(db)
	}

	return nil
}

// Scan stored SpoilerLogs and update all Seeds in the database
func updateAllSeedsFromStoredSpoilerLogs(db *gorm.DB) error {
	spoilerLogFiles := make([]randoseed.SpoilerLogFile, 0)
	if err := db.Preload("Seed").Find(&spoilerLogFiles).Error; err != nil {
		return err
	} else {
		for _, entry := range spoilerLogFiles {

			if entry.Seed == nil {
				log.Printf("migration: SpoilerLogFile ID=%d has nil Seed (seed_id=%d); skipping entry\n", entry.ID, entry.SeedID)
				continue
			}

			fileHash := entry.Seed.FileHash

			seed, seedErr := randoseed.GetByFileHashWithRelationships(db, fileHash)
			if seedErr != nil {
				return seedErr
			}

			spoilerLog, spoilerErr := randoseed.GetSpoilerLogFromDBRecord(&entry)
			if spoilerErr != nil {
				return spoilerErr
			} else if spoilerLog == nil {
				return fmt.Errorf("unable to get SpoilerLog from %s", fileHash)
			}

			spoilerLog.UpdateDatabaseSeed(seed)

			// Ensure we update existing RawSettings instead of inserting a new one
			if seed.RawSettings != nil {
				var existingRaw randoseed.RawSettings
				if err := db.Where("seed_id = ?", seed.ID).First(&existingRaw).Error; err == nil {
					seed.RawSettings.ID = existingRaw.ID
				}
			}

			if err := db.Save(seed).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
