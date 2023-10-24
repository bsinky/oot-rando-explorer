package randoseed

import "time"

type DBSeed struct {
	Id          int64
	UploadTime  time.Time
	Seed        string
	Version     string
	FileHash    string
	Logic       string
	Shopsanity  string
	Tokensanity string
	Scrubsanity string
	RawSettings string
}

func (seed DBSeed) FormattedUploadTime() string {
	return seed.UploadTime.Format(time.RFC1123)
}

func MakeDatabaseRecord(s SpoilerLog, rawSettings string) DBSeed {
	return DBSeed{
		UploadTime:  time.Now(),
		Seed:        s.Seed,
		Version:     s.Version,
		FileHash:    s.FileHashString(),
		Logic:       s.Settings.Logic,
		Shopsanity:  s.Settings.Shopsanity,
		Tokensanity: s.Settings.Tokensanity,
		Scrubsanity: s.Settings.Scrubsanity,
		RawSettings: rawSettings,
	}
}
