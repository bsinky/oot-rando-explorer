package randoseed

type SeedRank struct {
	ID         int64
	UserID     int64 `validate:"required"`
	DBSeedID   int64 `validate:"required"`
	Difficulty uint8 `validate:"required,gte=0,lte=5" form:"difficulty"`
	Fun        uint8 `validate:"required,gte=0,lte=5" form:"fun"`
}

type AvgSeedRank struct {
	DBSeedID   int64
	TotalVotes int64
	Difficulty float64
	Fun        float64
}
