package randoseed

import (
	"time"

	"gorm.io/gorm"
)

type Seed struct {
	gorm.Model
	Seed          string
	Version       string
	FileHash      string
	Logic         string
	Shopsanity    string
	Tokensanity   string
	Scrubsanity   string
	MQDungeons    string
	ItemPool      string
	EntranceRando string
	RawSettings   string
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

func MostRecent(db *gorm.DB, n int) ([]Seed, error) {
	seeds := make([]Seed, 0, n)
	if err := db.Order("ID DESC").Limit(n).Find(&seeds).Error; err != nil {
		return nil, err
	}

	return seeds, nil
}
