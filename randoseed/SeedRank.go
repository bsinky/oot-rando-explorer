package randoseed

import (
	"errors"

	"github.com/bsinky/sohrando/authentication"
	"gorm.io/gorm"
)

type SeedRank struct {
	ID         uint
	UserID     uint `validate:"required"`
	User       *authentication.User
	SeedID     uint `validate:"required"`
	Seed       *Seed
	Difficulty uint8 `validate:"required,gte=0,lte=5" form:"difficulty"`
	Fun        uint8 `validate:"required,gte=0,lte=5" form:"fun"`
}

type AvgSeedRank struct {
	SeedID     uint
	TotalVotes uint
	Difficulty float64
	Fun        float64
}

func GetUserRank(db *gorm.DB, seedID uint, userID uint) (*SeedRank, error) {
	var seedRank SeedRank
	if err := db.Where(&SeedRank{UserID: userID, SeedID: seedID}).First(&seedRank).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &seedRank, nil
}

func GetAverageRank(db *gorm.DB, seedID uint) (*AvgSeedRank, error) {
	var avgSeedRank AvgSeedRank
	res := db.Raw(`SELECT
	    seed_ranks.seed_id,
		COUNT(seed_ranks.id) AS total_votes,
		AVG(seed_ranks.difficulty) AS difficulty,
		AVG(seed_ranks.fun) AS fun
	FROM seed_ranks
	WHERE seed_ranks.seed_id = ?
	GROUP BY seed_ranks.seed_id`, seedID).First(&avgSeedRank)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &avgSeedRank, nil
}
