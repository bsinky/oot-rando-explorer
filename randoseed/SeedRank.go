package randoseed

import (
	"errors"

	"github.com/bsinky/sohrando/authentication"
	"gorm.io/gorm"
)

type SeedRank struct {
	ID         uint
	UserID     uint `binding:"required" gorm:"index:idx_seed_id_user_id,priority:2"`
	User       *authentication.User
	SeedID     uint `binding:"required" gorm:"index:idx_seed_id_user_id,priority:1"`
	Seed       *Seed
	Difficulty uint8 `binding:"required,gte=0,lte=5" form:"difficulty"`
	Fun        uint8 `binding:"required,gte=0,lte=5" form:"fun"`
}

type AvgSeedRank struct {
	ID         uint
	SeedID     uint `gorm:"uniqueIndex,index:idx_difficulty,priority:2,index:idx_fun,priority2"`
	Seed       *Seed
	TotalVotes uint
	Difficulty float64 `gorm:"index:idx_difficulty,priority:1"`
	Fun        float64 `gorm:"index:idx_fun,priority:1"`
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

func UpdateAverageRank(db *gorm.DB, seedID uint) (*AvgSeedRank, error) {
	var calcAvgRank AvgSeedRank
	// TODO: replace Raw usage to make this portable to other dbs
	res := db.Raw(`SELECT
	    seed_ranks.seed_id,
		COUNT(seed_ranks.id) AS total_votes,
		AVG(seed_ranks.difficulty) AS difficulty,
		AVG(seed_ranks.fun) AS fun
	FROM seed_ranks
	WHERE seed_ranks.seed_id = ?
	GROUP BY seed_ranks.seed_id`, seedID).First(&calcAvgRank)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	avgSeedRank := AvgSeedRank{}
	if err := db.Where(&AvgSeedRank{SeedID: seedID}).Assign(AvgSeedRank{
		SeedID:     calcAvgRank.SeedID,
		TotalVotes: calcAvgRank.TotalVotes,
		Difficulty: calcAvgRank.Difficulty,
		Fun:        calcAvgRank.Fun}).FirstOrInit(&avgSeedRank).Error; err != nil {
		return nil, err
	}

	if err := db.Save(&avgSeedRank).Error; err != nil {
		return nil, err
	}

	return &avgSeedRank, nil
}

func topAvgSeedRanks(db *gorm.DB, n int, sortField string, sortDirection string, prevValue *float64, prevID *uint) ([]AvgSeedRank, error) {
	seeds := make([]AvgSeedRank, 0, n)

	query := db.Preload("Seed").Joins("INNER JOIN seeds ON avg_seed_ranks.seed_id = seeds.id AND seeds.deleted_at IS NULL").Order(sortField + " " + sortDirection).Order("avg_seed_ranks.id ASC").Limit(n)

	if prevValue != nil && prevID != nil {
		var operation string
		if sortDirection == "DESC" {
			operation = "<"
		} else {
			operation = ">"
		}
		if err := query.Where(sortField+" "+operation+" ?", prevValue).Or(sortField+" = ? AND avg_seed_ranks.id > ?", prevValue, prevID).Find(&seeds).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
	} else {
		if err := query.Find(&seeds).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
	}

	return seeds, nil
}

func EasiestSeeds(db *gorm.DB, n int, prevValue *float64, prevID *uint) ([]AvgSeedRank, error) {
	return topAvgSeedRanks(db, n, "difficulty", "ASC", prevValue, prevID)
}

func HardestSeeds(db *gorm.DB, n int, prevValue *float64, prevID *uint) ([]AvgSeedRank, error) {
	return topAvgSeedRanks(db, n, "difficulty", "DESC", prevValue, prevID)
}

func MostFunSeeds(db *gorm.DB, n int, prevValue *float64, prevID *uint) ([]AvgSeedRank, error) {
	return topAvgSeedRanks(db, n, "fun", "DESC", prevValue, prevID)
}
