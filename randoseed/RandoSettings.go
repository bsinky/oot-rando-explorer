package randoseed

import (
	"github.com/bsinky/sohrando/randoseed/entrancerando"
	"github.com/bsinky/sohrando/randoseed/itempool"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/mqdungeons"
	"github.com/bsinky/sohrando/randoseed/scrubsanity"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/tokensanity"
)

type RandoSettings struct {
	Logic         string `json:"Logic Options:Logic"`
	Shopsanity    string `json:"Shuffle Settings:Shopsanity"`
	Tokensanity   string `json:"Shuffle Settings:Tokensanity"`
	Scrubsanity   string `json:"Shuffle Settings:Scrub Shuffle"`
	MQDungeons    string `json:"World Settings:MQ Dungeon Count"`
	ItemPool      string `json:"Item Pool Settings:Item Pool"`
	EntranceRando string `json:"World Settings:Shuffle Entrances"`
}

func (s *RandoSettings) LogicOrDefault() logic.Logic {
	if v, err := logic.FromDisplayName(s.Logic); err != nil {
		return logic.Glitchless
	} else {
		return v
	}
}

func (s *RandoSettings) ShopsanityOrDefault() shopsanity.Shopsanity {
	if v, err := shopsanity.FromDisplayName(s.Shopsanity); err != nil {
		return shopsanity.Off
	} else {
		return v
	}
}

func (s *RandoSettings) TokensanityOrDefault() tokensanity.Tokensanity {
	if v, err := tokensanity.FromDisplayName(s.Tokensanity); err != nil {
		return tokensanity.Off
	} else {
		return v
	}
}

func (s *RandoSettings) ScrubsanityOrDefault() scrubsanity.Scrubsanity {
	if v, err := scrubsanity.FromDisplayName(s.Scrubsanity); err != nil {
		return scrubsanity.Off
	} else {
		return v
	}
}

func (s *RandoSettings) MQDungeonsOrDefault() mqdungeons.MQDungeons {
	if v, err := mqdungeons.FromDisplayName(s.MQDungeons); err != nil {
		return mqdungeons.Zero
	} else {
		return v
	}
}

func (s *RandoSettings) ItemPoolOrDefault() itempool.ItemPool {
	if v, err := itempool.FromDisplayName(s.ItemPool); err != nil {
		return itempool.Balanced
	} else {
		return v
	}
}

func (s *RandoSettings) EntranceRandoOrDefault() entrancerando.EntranceRando {
	if v, err := entrancerando.FromDisplayName(s.EntranceRando); err != nil {
		return entrancerando.Off
	} else {
		return v
	}
}
