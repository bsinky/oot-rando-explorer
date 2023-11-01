package randoseed

import (
	"time"

	"gorm.io/gorm"
)

type Seed struct {
	gorm.Model
	Seed        string
	Version     string
	FileHash    string
	Logic       string
	Shopsanity  string
	Tokensanity string
	Scrubsanity string
	RawSettings string
}

// TODO: add more settings columns, things that would be useful to
// TODO: for filtering on in the future:
// TODO:   - MQ Dungeons
// TODO:   - Item Pool (Balanced, Scarce, etc.)
// TODO:   - Entrance Rando

func (seed Seed) FormattedUploadTime() string {
	return seed.CreatedAt.Format(time.RFC1123)
}

func MakeDatabaseRecord(s *SpoilerLog) Seed {
	return Seed{
		Seed:        s.Seed,
		Version:     s.Version,
		FileHash:    s.FileHashString(),
		Logic:       s.Settings.Logic,
		Shopsanity:  s.Settings.Shopsanity,
		Tokensanity: s.Settings.Tokensanity,
		Scrubsanity: s.Settings.Scrubsanity,
		RawSettings: s.RawSettings,
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
