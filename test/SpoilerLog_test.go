package main

import (
	"encoding/json"
	"ootrandoexplorer/site/randoseed"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func testReadingSpoilerLog(t *testing.T, filePath string) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	spoilerLog := randoseed.SpoilerLog{}
	jsonErr := json.Unmarshal(fileBytes, &spoilerLog)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	if err := validate.Struct(&spoilerLog); err != nil {
		errs := err.(validator.ValidationErrors)
		t.Fatalf(errs.Error())
	}
}

// TODO: every 5.1.4 seed I generate gives "unexpected end of input" JSON errors,
// TODO: but all the online validators I've tried say its valid. Something is causing
// TODO: Go not to read the whole file maybe? Not sure what's going on.
// Bradley Echo 5.1.4
func TestReadingBradleyEchoSpoilerLog(t *testing.T) {
	testReadingSpoilerLog(t, "37-07-80-16-59.json")
}

// Khan Bravo 6.1.1
func TestReadingKhanBravoSpoilerLog(t *testing.T) {
	testReadingSpoilerLog(t, "04-94-01-69-66.json")
}

// Spock Charlie 7.0.2
func TestReadingSpockCharlieSpoilerLog(t *testing.T) {
	testReadingSpoilerLog(t, "30-22-68-19-81.json")
}

// Sulu Bravo 7.1.1
func TestReadingSuluBravoSpoilerLog(t *testing.T) {
	testReadingSpoilerLog(t, "27-46-32-77-65.json")
}
