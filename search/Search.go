package search

import (
	"errors"
	"slices"
	"sync"

	"github.com/bsinky/sohrando/randoseed"
	"gorm.io/gorm"
)

type SearchFilter struct {
	FieldName string
	Label     string
	Options   []string
}

func (s *SearchFilter) IsValidOption(v string) bool {
	for _, opt := range s.Options {
		if opt == v {
			return true
		}
	}
	return false
}

type SearchFilterValue struct {
	FieldName string
	Value     string
}

type Result struct {
	Seeds []randoseed.Seed
}

type ResultModel struct {
	Result Result
}

var versionsMostRecentFirst = sync.OnceValue(func() []string {
	versionsCopy := make([]string, len(randoseed.Versions))
	copy(versionsCopy, randoseed.Versions)
	slices.Reverse(versionsCopy)
	return versionsCopy
})

func AllFilters(db *gorm.DB) (map[string]*SearchFilter, error) {
	filterMap := map[string]*SearchFilter{
		"Version": {
			Label:   "Version",
			Options: versionsMostRecentFirst(),
		},
		"Logic": {
			Label: "Logic",
			Options: []string{
				"Glitchless",
				"Glitched",
				"No Logic",
				"Vanilla",
			},
		},
		"Shopsanity": {
			Label: "Shopsanity",
			Options: []string{
				"Off",
				"0 Items",
				"1 Item",
				"2 Items",
				"3 Items",
				"4 Items",
				"Random",
			},
		},
		"Tokensanity": {
			Label: "Tokensanity",
			Options: []string{
				"Off",
				"Dungeons",
				"Overworld",
				"All Tokens",
			},
		},
		"Scrubsanity": {
			Label: "Scrubsanity",
			Options: []string{
				"Off",
				"Affordable",
				"Expensive",
				"Random Prices",
			},
		},
		"MQDungeons": {
			Label: "MQ Dungeons",
			Options: []string{
				"Selection",
				"Random",
				"0",
				"1",
				"2",
				"3",
				"4",
				"5",
				"6",
				"7",
				"8",
				"9",
				"10",
				"11",
				"12",
			},
		},
		"ItemPool": {
			Label: "Item Pool",
			Options: []string{
				"Plentiful",
				"Balanced",
				"Scarce",
				"Minimal",
			},
		},
		"EntranceRando": {
			Label: "Entrance Rando",
			Options: []string{
				"Off",
				"On",
			},
		},
	}

	for k, v := range filterMap {
		if v.FieldName == "" {
			v.FieldName = k
		}
	}

	return filterMap, nil
}

func RunSearch(db *gorm.DB, filters []SearchFilterValue) (*Result, error) {
	var seeds []randoseed.Seed

	seedQuery := db.Model(&randoseed.Seed{})

	for _, filter := range filters {
		dbColName := db.NamingStrategy.ColumnName("seeds", filter.FieldName)
		seedQuery = seedQuery.Where(dbColName+" = ?", filter.Value)
	}

	if err := seedQuery.Limit(10).Find(&seeds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &Result{}, nil
		}
		return nil, err
	}
	return &Result{
		Seeds: seeds,
	}, nil
}
