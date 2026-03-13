package randoseed

import (
	"github.com/bsinky/sohrando/randoseed/entrancerando"
	"github.com/bsinky/sohrando/randoseed/itempool"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/mqdungeons"
	"github.com/bsinky/sohrando/randoseed/scrubsanity"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/startingage"
	"github.com/bsinky/sohrando/randoseed/tokensanity"
)

// Before 9.0.0, SpoilerLog JSON had a different (Legacy) format
type RandoSettings struct {
	Logic               string `json:"Logic"`
	LegacyLogic         string `json:"Logic Options:Logic,omitempty"`
	Shopsanity          string `json:"Shop Shuffle"`
	LegacyShopsanity    string `json:"Shuffle Settings:Shopsanity,omitempty"`
	Tokensanity         string `json:"Token Shuffle"`
	LegacyTokensanity   string `json:"Shuffle Settings:Tokensanity,omitempty"`
	Scrubsanity         string `json:"Scrub Shuffle"`
	LegacyScrubsanity   string `json:"Shuffle Settings:Scrub Shuffle,omitempty"`
	MQDungeons          string `json:"MQ Dungeon Count"`
	LegacyMQDungeons    string `json:"World Settings:MQ Dungeon Count,omitempty"`
	ItemPool            string `json:"Item Pool"`
	LegacyItemPool      string `json:"Item Pool Settings:Item Poo,omitempty"`
	EntranceRando       string `json:"Shuffle Entrances"`
	LegacyEntranceRando string `json:"World Settings:Shuffle Entrances,omitempty"`
	StartingAge         string `json:"Starting Age"`
	LegacyStartingAge   string `json:"World Settings:Starting Age,omitempty"`
}

func (s *RandoSettings) LogicOrDefault() logic.Logic {
	logicToUse := s.Logic
	if logicToUse == "" && s.LegacyLogic != "" {
		logicToUse = s.LegacyLogic
	}
	if v, err := logic.FromDisplayName(logicToUse); err != nil {
		return logic.Glitchless
	} else {
		return v
	}
}

func (s *RandoSettings) ShopsanityOrDefault() shopsanity.Shopsanity {
	shopsanityToUse := s.Shopsanity
	if shopsanityToUse == "" && s.LegacyShopsanity != "" {
		shopsanityToUse = s.LegacyShopsanity
	}
	if v, err := shopsanity.FromDisplayName(shopsanityToUse); err != nil {
		return shopsanity.Off
	} else {
		return v
	}
}

func (s *RandoSettings) TokensanityOrDefault() tokensanity.Tokensanity {
	tokensanityToUse := s.Tokensanity
	if tokensanityToUse == "" && s.LegacyTokensanity != "" {
		tokensanityToUse = s.LegacyTokensanity
	}
	if v, err := tokensanity.FromDisplayName(tokensanityToUse); err != nil {
		return tokensanity.Off
	} else {
		return v
	}
}

func (s *RandoSettings) ScrubsanityOrDefault() scrubsanity.Scrubsanity {
	scrubsanityToUse := s.Scrubsanity
	if scrubsanityToUse == "" && s.LegacyScrubsanity != "" {
		scrubsanityToUse = s.LegacyScrubsanity
	}
	if v, err := scrubsanity.FromDisplayName(scrubsanityToUse); err != nil {
		return scrubsanity.Off
	} else {
		return v
	}
}

func (s *RandoSettings) MQDungeonsOrDefault() mqdungeons.MQDungeons {
	mqDungeonsToUse := s.MQDungeons
	if mqDungeonsToUse == "" && s.LegacyMQDungeons != "" {
		mqDungeonsToUse = s.LegacyMQDungeons
	}
	if v, err := mqdungeons.FromDisplayName(mqDungeonsToUse); err != nil {
		return mqdungeons.Zero
	} else {
		return v
	}
}

func (s *RandoSettings) ItemPoolOrDefault() itempool.ItemPool {
	itemPoolToUse := s.ItemPool
	if itemPoolToUse == "" && s.LegacyItemPool != "" {
		itemPoolToUse = s.LegacyItemPool
	}
	if v, err := itempool.FromDisplayName(itemPoolToUse); err != nil {
		return itempool.Balanced
	} else {
		return v
	}
}

func (s *RandoSettings) EntranceRandoOrDefault() entrancerando.EntranceRando {
	settingToUse := s.EntranceRando
	if settingToUse == "" && s.LegacyEntranceRando != "" {
		settingToUse = s.LegacyEntranceRando
	}
	if v, err := entrancerando.FromDisplayName(settingToUse); err != nil {
		return entrancerando.Off
	} else {
		return v
	}
}

func (s *RandoSettings) StartingAgeOrDefault() startingage.StartingAge {
	ageToUse := s.StartingAge
	if ageToUse == "" && s.LegacyStartingAge != "" {
		ageToUse = s.LegacyStartingAge
	}
	if v, err := startingage.FromDisplayName(ageToUse); err != nil {
		return startingage.Random
	} else {
		return v
	}
}
