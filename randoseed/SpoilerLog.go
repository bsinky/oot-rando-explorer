package randoseed

import (
	"strconv"
	"strings"
)

type SpoilerLog struct {
	Seed     string        `json:"_seed"`
	Version  string        `json:"_version"`
	FileHash []int         `json:"file_hash"`
	Settings RandoSettings `json:"settings"`
}

// func (s SpoilerLog) UnmarshalJSON(data []byte) error {
// 	spoilerLog := SpoilerLog{}
// 	strMap := make(map[string]string)
// 	err := json.Unmarshal(data, &strMap)
// 	if err != nil {
// 		return err
// 	}
// 	// TODO: handle settings explicitly somehow
// 	spoilerLog.Seed = strMap["_seed"]
// 	spoilerLog.Version = strMap["_version"]
// 	// spoilerLog.FileHash = strMap["file_hash"]
// 	spoilerLog.Version = strMap["_version"]

// 	return nil
// }

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
