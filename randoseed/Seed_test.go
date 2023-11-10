package randoseed

import (
	"strings"
	"testing"
)

func TestVersions(t *testing.T) {
	if len(Versions) == 0 {
		t.Fatalf("Expected >0 Versions, got %d", len(Versions))
	}

	for _, version := range Versions {
		strippedVersion := strings.TrimSpace(version)
		if strippedVersion == "" {
			t.Fatalf("Versions should not contain empty or whitespace-only entries")
		} else if len(version) != len(strippedVersion) {
			t.Fatalf("Version should contain leading or trailing whitespace: %s", version)
		}
	}
}
