package randoseed

import (
	_ "embed"
	"strings"
	"time"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/randoseed/entrancerando"
	"github.com/bsinky/sohrando/randoseed/itempool"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/mqdungeons"
	"github.com/bsinky/sohrando/randoseed/scrubsanity"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/tokensanity"
	"github.com/go-playground/validator/v10"
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
	val := fl.Field().String()
	for _, version := range Versions {
		if val == version {
			return true
		}
	}

	return false
}

// TODO: possibly move Version to a separate table to better normalize data and save storage?
type Seed struct {
	gorm.Model
	Seed            string
	VersionID       uint `gorm:"index" validate:"required,validVersion"`
	Version         *Version
	FileHash        string                      `gorm:"uniqueIndex" validate:"required"`
	Logic           logic.Logic                 `gorm:"index"`
	Shopsanity      shopsanity.Shopsanity       `gorm:"index"`
	Tokensanity     tokensanity.Tokensanity     `gorm:"index"`
	Scrubsanity     scrubsanity.Scrubsanity     `gorm:"index"`
	MQDungeons      mqdungeons.MQDungeons       `gorm:"index"`
	ItemPool        itempool.ItemPool           `gorm:"index"`
	EntranceRando   entrancerando.EntranceRando `gorm:"index"`
	RawSettings     *RawSettings
	User            *authentication.User `gorm:"foreignKey:UserIDUploader"`
	UserIDUploader  uint                 `gorm:"index" validate:"required"`
	UploaderComment string               `validate:"len=500" form:"uploaderComment"`
}

type Version struct {
	ID   uint
	Name string
}

type RawSettings struct {
	ID           uint
	SettingsJSON string `validate:"len=10000"`
	SeedID       uint   `gorm:"uniqueIndex" validate:"required"`
}

type Setting struct {
	Label string
	Value string
}

func (seed Seed) FormattedUploadTime() string {
	return seed.CreatedAt.Format(time.RFC1123)
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
