package randoseed

import (
	_ "embed"
	"errors"
	"strings"
	"time"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/randoseed/entrancerando"
	"github.com/bsinky/sohrando/randoseed/itempool"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/mqdungeons"
	"github.com/bsinky/sohrando/randoseed/scrubsanity"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/startingage"
	"github.com/bsinky/sohrando/randoseed/tokensanity"
	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

//go:embed versions.txt
var versionsString string

var VersionNames []string = strings.Split(strings.TrimSpace(versionsString), "\n")
var Versions map[uint]string
var VersionIDs map[string]uint

func VersionsMostRecentFirst(db *gorm.DB) ([]Version, error) {
	options := make([]Version, 0, len(VersionNames))

	if err := db.Order("id DESC").Find(&options).Error; err != nil {
		return nil, err
	}

	return options, nil
}

func InitVersionCache(db *gorm.DB) error {
	Versions = make(map[uint]string)
	VersionIDs = make(map[string]uint)

	versions := make([]Version, 0)
	if err := db.Find(&versions).Error; err != nil {
		return err
	}

	addToCache := func(v *Version) {
		Versions[v.ID] = v.Name
		VersionIDs[v.Name] = v.ID
	}

	for _, v := range versions {
		addToCache(&v)
	}

	for _, versionFromFile := range VersionNames {
		if _, ok := VersionIDs[versionFromFile]; ok {
			continue
		}

		// not in DB, create one
		newVersion := &Version{
			Name: versionFromFile,
		}

		if err := db.Save(newVersion).Error; err != nil {
			return err
		}

		addToCache(newVersion)
	}

	return nil
}

func RegisterValidation(v *validator.Validate) {
	v.RegisterValidation("validVersion", validateVersion)
}

func validateVersion(fl validator.FieldLevel) bool {
	val := fl.Field().Uint()
	_, ok := Versions[uint(val)]

	return ok
}

type Seed struct {
	//--gorm.Model fields
	//-- not embedding since we want DeletedAt to be part of a composite index
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index:idx_file_hash_deleted_at,unique,priority:2"`
	//---
	Seed            string
	VersionID       uint `gorm:"index" binding:"validVersion"`
	Version         *Version
	FileHash        string                      `gorm:"index:idx_file_hash_deleted_at,unique,priority:1" binding:"required"`
	Logic           logic.Logic                 `gorm:"index"`
	Shopsanity      shopsanity.Shopsanity       `gorm:"index"`
	Tokensanity     tokensanity.Tokensanity     `gorm:"index"`
	Scrubsanity     scrubsanity.Scrubsanity     `gorm:"index"`
	MQDungeons      mqdungeons.MQDungeons       `gorm:"index"`
	ItemPool        itempool.ItemPool           `gorm:"index"`
	EntranceRando   entrancerando.EntranceRando `gorm:"index"`
	StartingAge     startingage.StartingAge     `gorm:"index"`
	RawSettings     *RawSettings
	User            *authentication.User `gorm:"foreignKey:UserIDUploader"`
	UserIDUploader  uint                 `gorm:"index" binding:"required"`
	UploaderComment string               `binding:"max=500" form:"uploaderComment"`
}

type Version struct {
	ID   uint
	Name string
}

type RawSettings struct {
	ID           uint
	SettingsJSON string `binding:"max=10000"`
	SeedID       uint   `gorm:"uniqueIndex" validate:"required"`
}

type SpoilerLogFile struct {
	gorm.Model
	SeedID         uint `gorm:"index" binding:"required"`
	Seed           *Seed
	SpoilerLogJSON datatypes.JSON
}

type Setting struct {
	Label string
	Value string
}

func (seed Seed) FormattedUploadTime() string {
	return seed.CreatedAt.Format(time.RFC3339)
}

func (seed *Seed) CachedVersion() string {
	if version, ok := Versions[seed.VersionID]; ok {
		return version
	}

	return "Unknown Version"
}

func (seed Seed) Settings() []Setting {
	return []Setting{
		{
			Label: "Logic",
			Value: seed.Logic.DisplayName(),
		},
		{
			Label: "Shopsanity",
			Value: seed.Shopsanity.DisplayName(),
		},
		{
			Label: "Tokensanity",
			Value: seed.Tokensanity.DisplayName(),
		},
		{
			Label: "Scrubsanity",
			Value: seed.Scrubsanity.DisplayName(),
		},
		{
			Label: "MQ Dungeons",
			Value: seed.MQDungeons.DisplayName(),
		},
		{
			Label: "Item Pool",
			Value: seed.ItemPool.DisplayName(),
		},
		{
			Label: "Entrance Rando",
			Value: seed.EntranceRando.DisplayName(),
		},
		{
			Label: "Starting Age",
			Value: seed.StartingAge.DisplayName(),
		},
	}
}

func GetByFileHash(db *gorm.DB, fileHash string) (*Seed, error) {
	var seed Seed
	if err := db.First(&seed, "file_hash = ?", fileHash).Error; err != nil {
		return nil, err
	}

	return &seed, nil
}

func GetByFileHashWithRelationships(db *gorm.DB, fileHash string) (*Seed, error) {
	var seed Seed
	if err := db.Preload("RawSettings").Preload("User").First(&seed, "file_hash = ?", fileHash).Error; err != nil {
		return nil, err
	}

	return &seed, nil
}

func MostRecent(db *gorm.DB, n int) ([]Seed, error) {
	seeds := make([]Seed, 0, n)
	if err := db.Order("ID DESC").Limit(n).Find(&seeds).Error; err != nil {
		return nil, err
	}

	return seeds, nil
}

func UserUploadedSeeds(db *gorm.DB, userID uint) ([]Seed, error) {
	batchSize := 10
	seeds := make([]Seed, 0, batchSize)
	if err := db.Order("ID DESC").Limit(batchSize).Find(&seeds, &Seed{UserIDUploader: userID}).Error; err != nil {
		return nil, err
	}

	return seeds, nil
}

func GetSpoilerLogFile(db *gorm.DB, seedID uint) (*SpoilerLogFile, error) {
	file := &SpoilerLogFile{}
	if err := db.First(file, &SpoilerLogFile{SeedID: seedID}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return file, nil
}
