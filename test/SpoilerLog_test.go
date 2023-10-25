package main

import (
	"encoding/json"
	"ootrandoexplorer/site/randoseed"
	"os"
	"testing"
)

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

	if spoilerLog.Version == "" {
		t.Fatalf(`Version was empty after deserialization`)
	}

	if spoilerLog.Seed == "" {
		t.Fatalf(`Seed was empty after deserialization`)
	}
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
