package main

import (
	"ootrandoexplorer/site/randoseed"
	"testing"
)

func TestFileHashString(t *testing.T) {
	s := randoseed.SpoilerLog{
		Seed:     "",
		Version:  "",
		FileHash: []uint{1, 2, 3, 4, 5},
		Settings: randoseed.RandoSettings{},
	}
	want := "01-02-03-04-05"
	actual := s.FileHashString()
	if want != actual {
		t.Fatalf(`s.FileHashString() = %q, want %q`, actual, want)
	}
}
