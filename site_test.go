package main

import (
	"testing"
)

func TestFileHashString(t *testing.T) {
	s := SpoilerLog{
		Seed:     "",
		Version:  "",
		FileHash: []int{1, 2, 3, 4, 5},
		Settings: RandoSettings{},
	}
	want := "01-02-03-04-05"
	actual := s.FileHashString()
	if want != actual {
		t.Fatalf(`s.FileHashString() = %q, want %q`, actual, want)
	}
}
