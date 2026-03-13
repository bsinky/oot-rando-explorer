package randoseed

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/bsinky/sohrando/randoseed/entrancerando"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/mqdungeons"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/startingage"
	"github.com/bsinky/sohrando/randoseed/tokensanity"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func testReadingSpoilerLog(t *testing.T, filePath string) *SpoilerLog {
	fileBytes, err := os.ReadFile(path.Join("..", "test", filePath))
	if err != nil {
		t.Fatal(err)
	}

	spoilerLog := SpoilerLog{}
	jsonErr := json.Unmarshal(fileBytes, &spoilerLog)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	if err := validate.Struct(&spoilerLog); err != nil {
		errs := err.(validator.ValidationErrors)
		t.Fatalf("%s", errs.Error())
	}

	return &spoilerLog
}

func testRandoSettingsParsing(t *testing.T, filePath string, versionName string,
	expectedLogic logic.Logic, expectedTokensanity tokensanity.Tokensanity,
	expectedShopsanity shopsanity.Shopsanity, expectedMQDungeons mqdungeons.MQDungeons,
	expectedEntranceRando entrancerando.EntranceRando,
	expectedStartingAge startingage.StartingAge) {
	spoilerLog := testReadingSpoilerLog(t, filePath)

	actualLogic := spoilerLog.Settings.LogicOrDefault()
	assert.Equalf(t, expectedLogic, actualLogic,
		"%s Logic parsing failed", versionName)

	actualTokensanity := spoilerLog.Settings.TokensanityOrDefault()
	assert.Equal(t, expectedTokensanity, actualTokensanity,
		"%s Tokensanity parsing failed", versionName)

	actualShopsanity := spoilerLog.Settings.ShopsanityOrDefault()
	assert.Equalf(t, expectedShopsanity, actualShopsanity,
		"%s Shopsanity parsing failed", versionName)

	actualMQDungeons := spoilerLog.Settings.MQDungeonsOrDefault()
	assert.Equal(t, expectedMQDungeons, actualMQDungeons,
		"%s MQ Dungeons parsing failed", versionName)

	actualEntranceRando := spoilerLog.Settings.EntranceRandoOrDefault()
	assert.Equal(t, expectedEntranceRando, actualEntranceRando,
		"%s Entrance Rando parsing failed", versionName)

	actualStartingAge := spoilerLog.Settings.StartingAgeOrDefault()
	assert.Equal(t, expectedStartingAge, actualStartingAge,
		"%s Starting Age parsing failed", versionName)
}

func TestFileHashString(t *testing.T) {
	s := SpoilerLog{
		Seed:     "",
		Version:  "",
		FileHash: []uint{1, 2, 3, 4, 5},
		Settings: RandoSettings{},
	}
	want := "01-02-03-04-05"
	actual := s.FileHashString()
	if want != actual {
		t.Fatalf(`s.FileHashString() = %q, want %q`, actual, want)
	}
}

// TODO: every 5.1.4 seed I generate gives "unexpected end of input" JSON errors,
// TODO: but all the online validators I've tried say its valid. Something is causing
// TODO: Go not to read the whole file maybe? Not sure what's going on.
// Bradley Echo 5.1.4
// func TestReadingBradleyEchoSpoilerLog(t *testing.T) {
// 	t.Parallel()
// 	testReadingSpoilerLog(t, "37-07-80-16-59.json")
// }

// Khan Bravo 6.1.1
func TestReadingKhanBravoSpoilerLog(t *testing.T) {
	t.Parallel()
	testReadingSpoilerLog(t, "04-94-01-69-66.json")
}

// Spock Charlie 7.0.2
func TestReadingSpockCharlieSpoilerLog(t *testing.T) {
	t.Parallel()
	testReadingSpoilerLog(t, "30-22-68-19-81.json")
}

// Sulu Bravo 7.1.1
const SuluBravoFileName = "27-46-32-77-65.json"

func TestReadingSuluBravoSpoilerLog(t *testing.T) {
	t.Parallel()
	testReadingSpoilerLog(t, SuluBravoFileName)
}

func TestSuluBravoSpoilerLogSettingsReadProperly(t *testing.T) {
	t.Parallel()
	testRandoSettingsParsing(t, SuluBravoFileName, "Sulu Bravo",
		logic.Glitchless, tokensanity.AllTokens,
		shopsanity.Random, mqdungeons.Random, entrancerando.Off,
		startingage.Child)
}

const CopperBravoFileName = "02-69-65-39-41.json"

// Copper Bravo 9.1.2
func TestReadingCopperBravoSpoilerLog(t *testing.T) {
	t.Parallel()
	testReadingSpoilerLog(t, CopperBravoFileName)
}

func TestCopperBravoSpoilerLogSettingsReadProperly(t *testing.T) {
	t.Parallel()
	testRandoSettingsParsing(t, CopperBravoFileName, "Copper Bravo",
		logic.Glitchless, tokensanity.AllTokens,
		shopsanity.Random, mqdungeons.Zero, entrancerando.On,
		startingage.Random)
}
