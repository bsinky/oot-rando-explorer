package randoseed

type RandoSettings struct {
	Logic         string `json:"Logic Options:Logic"`
	Shopsanity    string `json:"Shuffle Settings:Shopsanity"`
	Tokensanity   string `json:"Shuffle Settings:Tokensanity"`
	Scrubsanity   string `json:"Shuffle Settings:Scrub Shuffle"`
	MQDungeons    string `json:"World Settings:MQ Dungeon Count"`
	ItemPool      string `json:"Item Pool Settings:Item Pool"`
	EntranceRando string `json:"World Settings:Shuffle Entrances"`
}
