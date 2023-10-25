package randoseed

import (
	"encoding/json"
	"strconv"
	"strings"
)

// TODO: unify this with the DBSeed struct to just have a single struct representation of the seed
// TODO: then RandoSettings struct could also go away
type SpoilerLog struct {
	Seed        string
	Version     string
	FileHash    []int
	Settings    RandoSettings
	RawSettings string
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
		hashString.WriteString(strconv.Itoa((s.FileHash[i])))
		hashString.WriteString("-")
	}
	ret := hashString.String()
	return ret[:len(ret)-1] // remove trailing "-"
}
