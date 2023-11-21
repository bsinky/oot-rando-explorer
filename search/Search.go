package search

import (
	"errors"
	"fmt"
	"sort"

	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/randoseed/entrancerando"
	"github.com/bsinky/sohrando/randoseed/itempool"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/mqdungeons"
	"github.com/bsinky/sohrando/randoseed/scrubsanity"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/tokensanity"
	"gorm.io/gorm"
)

type SearchFilter struct {
	FieldName string
	Label     string
	Options   []FilterOption
}

type FilterOption struct {
	Value string
	Label string
}

type OptionsProvider interface {
	ValueMap() map[int]string
}

func optionsFromProvider(provider OptionsProvider) []FilterOption {
	options := make([]FilterOption, 0)
	vMap := provider.ValueMap()
	orderedKeys := make([]int, 0, len(vMap))
	for k := range vMap {
		orderedKeys = append(orderedKeys, k)
	}
	sort.Slice(orderedKeys, func(i, j int) bool {
		return i < j
	})

	for k := range orderedKeys {
		v := vMap[k]
		options = append(options, FilterOption{
			Value: fmt.Sprint(k),
			Label: v,
		})
	}

	return options
}

func (s *SearchFilter) IsValidOption(v string) bool {
	for _, opt := range s.Options {
		if opt.Value == v {
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

var versionsMostRecentFirst = func(db *gorm.DB) ([]FilterOption, error) {
	options := make([]FilterOption, 0, len(randoseed.VersionNames))

	if versions, err := randoseed.VersionsMostRecentFirst(db); err != nil {
		return nil, err
	} else {
		for _, v := range versions {
			options = append(options, FilterOption{
				Value: fmt.Sprint(v.ID),
				Label: v.Name,
			})
		}
	}
	return options, nil
}

func AllFilters(db *gorm.DB) (map[string]*SearchFilter, error) {
	var versions []FilterOption
	var err error
	if versions, err = versionsMostRecentFirst(db); err != nil {
		return nil, err
	}

	filterMap := map[string]*SearchFilter{
		"VersionID": {
			FieldName: "VersionID",
			Label:     "Version",
			Options:   versions,
		},
		"Logic": {
			Label:   "Logic",
			Options: optionsFromProvider(logic.LogicEnum{}),
		},
		"Shopsanity": {
			Label:   "Shopsanity",
			Options: optionsFromProvider(shopsanity.ShopsanityEnum{}),
		},
		"Tokensanity": {
			Label:   "Tokensanity",
			Options: optionsFromProvider(tokensanity.TokensanityEnum{}),
		},
		"Scrubsanity": {
			Label:   "Scrubsanity",
			Options: optionsFromProvider(scrubsanity.ScrubsanityEnum{}),
		},
		"MQDungeons": {
			Label:   "MQ Dungeons",
			Options: optionsFromProvider(mqdungeons.MQDungeonsEnum{}),
		},
		"ItemPool": {
			Label:   "Item Pool",
			Options: optionsFromProvider(itempool.ItemPoolEnum{}),
		},
		"EntranceRando": {
			Label:   "Entrance Rando",
			Options: optionsFromProvider(entrancerando.EntranceRandoEnum{}),
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
