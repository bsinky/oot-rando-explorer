package randoseed

type SeedRank struct {
	Id         int64
	UserID     string `validate:"required"`
	DBSeedId   int64  `validate:"required"`
	Difficulty uint8  `validate:"required,gte=0,lte=5" form:"difficulty"`
	Fun        uint8  `validate:"required,gte=0,lte=5" form:"fun"`
}

type AvgSeedRank struct {
	DBSeedId   int64
	TotalVotes int64
	Difficulty float64
	Fun        float64
}
