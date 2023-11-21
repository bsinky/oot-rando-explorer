package migration

import (
	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/randoseed"
	"gorm.io/gorm"
)

type MigrationHooks interface {
	BeforeAutoMigrate(db *gorm.DB) error
	Migrate(db *gorm.DB, storageDir string) error
}

func runAllBeforeHooks(db *gorm.DB, migrations []MigrationHooks) error {
	for _, migration := range migrations {
		if err := migration.BeforeAutoMigrate(db); err != nil {
			return err
		}
	}

	return nil
}

func runAll(db *gorm.DB, storageDir string, migrations []MigrationHooks) error {
	for _, migration := range migrations {
		if err := migration.Migrate(db, storageDir); err != nil {
			return err
		}
	}

	return nil
}

func MigrateDB(db *gorm.DB, storageDir string) error {
	migrations := []MigrationHooks{
		&MigrateSeed{},
	}

	if err := runAllBeforeHooks(db, migrations); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&randoseed.Seed{},
		&randoseed.SeedRank{},
		&authentication.User{},
		&randoseed.RawSettings{},
		&randoseed.AvgSeedRank{},
		&randoseed.Version{}); err != nil {
		return err
	}

	if err := runAll(db, storageDir, migrations); err != nil {
		return err
	}

	return nil
}
