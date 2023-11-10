package randoseed

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
)

type SpoilerLog struct {
	Seed        string `validate:"required_with=Version"`
	Version     string `validate:"required_with=Seed"`
	FileHash    []uint `validate:"len=5"`
	Settings    RandoSettings
	RawSettings string `validate:"required"`
}

func (spoilerLog *SpoilerLog) UnmarshalJSON(data []byte) error {
	strMap := make(map[string]json.RawMessage)
	err := json.Unmarshal(data, &strMap)
	if err != nil {
		return err
	}

	var rawSeed json.RawMessage
	if seedVal, ok := strMap["_seed"]; ok {
		rawSeed = seedVal
	} else if seedVal, ok := strMap["seed"]; ok {
		rawSeed = seedVal
	}
	if err := json.Unmarshal(rawSeed, &spoilerLog.Seed); err != nil {
		return err
	}

	var rawVersion json.RawMessage
	if versionVal, ok := strMap["_version"]; ok {
		rawVersion = versionVal
	} else if versionVal, ok := strMap["version"]; ok {
		rawVersion = versionVal
	}
	if err := json.Unmarshal(rawVersion, &spoilerLog.Version); err != nil {
		return err
	}

	if fileHashVal, ok := strMap["file_hash"]; ok {
		if err := json.Unmarshal(fileHashVal, &spoilerLog.FileHash); err != nil {
			return err
		}
	}

	if settingsVal, ok := strMap["settings"]; ok {
		if err := json.Unmarshal(settingsVal, &spoilerLog.Settings); err != nil {
			return err
		}
		(*spoilerLog).RawSettings = string(settingsVal)
	}

	return nil
}

func (s SpoilerLog) FileHashString() string {
	hashString := strings.Builder{}
	for i := 0; i < len(s.FileHash); i++ {
		if s.FileHash[i] < 10 {
			hashString.WriteString("0")
		}
		hashString.WriteString(strconv.FormatUint(uint64(s.FileHash[i]), 10))
		hashString.WriteString("-")
	}
	ret := hashString.String()
	return ret[:len(ret)-1] // remove trailing "-"
}

func (s *SpoilerLog) CreateDatabaseSeed() *Seed {
	seed := &Seed{}
	s.UpdateDatabaseSeed(seed)
	return seed
}

func (s *SpoilerLog) UpdateDatabaseSeed(seed *Seed) {
	seed.Seed = s.Seed
	seed.Version = s.Version
	seed.FileHash = s.FileHashString()
	seed.Logic = s.Settings.Logic
	seed.Shopsanity = s.Settings.Shopsanity
	seed.Tokensanity = s.Settings.Tokensanity
	seed.Scrubsanity = s.Settings.Scrubsanity
	seed.MQDungeons = s.Settings.MQDungeons
	seed.ItemPool = s.Settings.ItemPool
	seed.EntranceRando = s.Settings.EntranceRando
	seed.RawSettings = &RawSettings{
		SettingsJSON: s.RawSettings,
	}
}

func GetSpoilerLogFromJsonFile(spoilerlogFile io.Reader) (*SpoilerLog, *bytes.Buffer, error) {
	spoilerLogBytes := new(bytes.Buffer)
	spoilerLogSize, err := io.Copy(spoilerLogBytes, spoilerlogFile)
	if err != nil || spoilerLogSize == 0 {
		return nil, nil, err
	}

	spoilerLog := &SpoilerLog{}
	jsonErr := json.Unmarshal(spoilerLogBytes.Bytes(), spoilerLog)
	if jsonErr != nil {
		return nil, nil, jsonErr
	}

	return spoilerLog, spoilerLogBytes, nil
}
