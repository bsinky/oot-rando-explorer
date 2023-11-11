package randoseed

import (
	_ "embed"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

//go:embed versions.txt
var versions string
var Versions []string = strings.Split(strings.TrimSpace(versions), "\n")

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
	Seed          string
	Version       string `gorm:"index" validate:"required,validVersion"`
	FileHash      string `gorm:"uniqueIndex" validate:"required"`
	Logic         string `gorm:"index"`
	Shopsanity    string `gorm:"index"`
	Tokensanity   string `gorm:"index"`
	Scrubsanity   string `gorm:"index"`
	MQDungeons    string `gorm:"index"`
	ItemPool      string `gorm:"index"`
	EntranceRando string `gorm:"index"`
	RawSettings   *RawSettings
}

type RawSettings struct {
	ID           uint
	SettingsJSON string `validate:"len:10000"`
	SeedID       uint   `gorm:"uniqueIndex" validate:"required"`
}

type Setting struct {
	Label string
	Value string
}

func (seed Seed) FormattedUploadTime() string {
	return seed.CreatedAt.Format(time.RFC1123)
}

func (seed Seed) Settings() []Setting {
	return []Setting{
		{
			Label: "Logic",
			Value: seed.Logic,
		},
		{
			Label: "Shopsanity",
			Value: seed.Shopsanity,
		},
		{
			Label: "Tokensanity",
			Value: seed.Tokensanity,
		},
		{
			Label: "Scrubsanity",
			Value: seed.Scrubsanity,
		},
		{
			Label: "MQ Dungeons",
			Value: seed.MQDungeons,
		},
		{
			Label: "Item Pool",
			Value: seed.ItemPool,
		},
		{
			Label: "Entrance Rando",
			Value: seed.EntranceRando,
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

func GetByFileHashWithRawSettings(db *gorm.DB, fileHash string) (*Seed, error) {
	var seed Seed
	if err := db.Preload("RawSettings").First(&seed, "file_hash = ?", fileHash).Error; err != nil {
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
