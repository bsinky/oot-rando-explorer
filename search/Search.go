package search

import (
	"errors"

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

type SearchModel struct {
	Filters map[string]*SearchFilter
}

type Result struct {
	Seeds []randoseed.Seed
}

type ResultModel struct {
	Result Result
}

func getVersionOptions(db *gorm.DB) ([]string, error) {
	var versions []string
	if err := db.Raw(`SELECT
		seeds.Version
	FROM seeds
	WHERE deleted_at IS NULL
	GROUP BY seeds.Version`).Scan(&versions).Error; err != nil {
		return nil, err
	}

	return versions, nil
}

func AllFilters(db *gorm.DB) (map[string]*SearchFilter, error) {
	versionOptions, err := getVersionOptions(db)
	if err != nil {
		return nil, err
	}
	filterMap := map[string]*SearchFilter{
		"Version": {
			Label:   "Version",
			Options: versionOptions,
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
