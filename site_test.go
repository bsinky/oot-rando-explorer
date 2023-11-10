package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/bsinky/sohrando/randoseed"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func bodyHasFragments(t *testing.T, body string, fragments []string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected body to contain '%s', got %s", fragment, body)
		}
	}
}

func getHasStatus(t *testing.T, db *gorm.DB, path string, status int) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, db, SpoilerSeedsDir)

	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		t.Errorf("got error: %s", err)
	}
	router.ServeHTTP(w, req)
	if status != w.Code {
		t.Errorf("expected response code %d, got %d", status, w.Code)
	}
	return w
}

func postHasStatus(t *testing.T, db *gorm.DB, path string, data *url.Values, status int) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, db, SpoilerSeedsDir)

	req, err := http.NewRequestWithContext(ctx, "POST", path, strings.NewReader(data.Encode()))
	if err != nil {
		t.Errorf("got error: %s", err)
	}
	router.ServeHTTP(w, req)
	if status != w.Code {
		t.Errorf("expected response code %d, got %d", status, w.Code)
	}
	return w
}

func createSeeds(t *testing.T, db *gorm.DB, count int) []*randoseed.Seed {
	seeds := []*randoseed.Seed{}
	t.Helper()
	for i := 0; i < count; i++ {
		s := randoseed.Seed{
			Version:     fmt.Sprintf("Test Case Version %d", i+1),
			FileHash:    fmt.Sprintf("%02d-%02d-%02d-%02d-%02d", i, i+1, i+2, i+3, i+4),
			Logic:       "Glitchless",
			Shopsanity:  "Off",
			Tokensanity: "Off",
			Scrubsanity: "Off",
			RawSettings: &randoseed.RawSettings{SettingsJSON: "{}"},
		}
		if err := db.Save(&s).Error; err != nil {
			t.Fatalf("error creating seed: %s", err)
		}
		seeds = append(seeds, &s)
	}
	return seeds
}

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

func TestEmptyDatabase(t *testing.T) {
	db := FreshDb(t)
	seeds, err := randoseed.MostRecent(db, 10)
	if err != nil {
		t.Fatalf("Error querying recent seeds from fresh db: %s", err)
	}
	if len(seeds) != 0 {
		t.Errorf("Expected 0 seeds, got %d", len(seeds))
	}
}

func TestMainPage(t *testing.T) {
	t.Parallel()

	db := FreshDb(t)

	createSeeds(t, db, 2)

	w := getHasStatus(t, db, "/", http.StatusOK)

	body := w.Body.String()

	if len(body) == 0 {
		t.Fatalf("expected response non-zero body length")
	}
	bodyHasFragments(t, body, []string{"Version 1", "Version 2"})
}

func TestSeedPage(t *testing.T) {
	t.Parallel()

	db := FreshDb(t)

	createSeeds(t, db, 3)

	w := getHasStatus(t, db, "/s/00-01-02-03-04", http.StatusOK)
	body := w.Body.String()
	bodyHasFragments(t, body, []string{"Test Case Version 1"})
}

func TestUploadSeed(t *testing.T) {
	db := FreshDb(t)

	routePath := "/uploadseed"
	status := http.StatusSeeOther

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, db, SpoilerSeedsDir)

	fileName := "04-94-01-69-66.json"
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()
		part, err := writer.CreateFormFile("spoilerlog", fileName)
		if err != nil {
			t.Error(err)
		}

		file, err := os.Open("test/" + fileName)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		_, err = io.Copy(part, file)
		if err != nil {
			t.Error(err)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", routePath, pr)
	if err != nil {
		t.Fatalf("got error: %s", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)
	if status != w.Code {
		t.Fatalf("expected response code %d, got %d", status, w.Code)
	}
	defer DeleteUploadedTestSeed(t, fileName)

	uploadedFile, err := os.Open(path.Join(SpoilerSeedsDir, fileName))
	if err != nil {
		t.Fatal(err)
	}
	defer uploadedFile.Close()
	uploadedFileStat, err := uploadedFile.Stat()
	if err != nil {
		t.Fatal(err)
	} else if uploadedFileStat.Size() == 0 {
		t.Fatal("Uploaded file was not the expected size")
	}
}

func TestVoteOnSeed(t *testing.T) {
	t.Parallel()

	db := FreshDb(t)

	createSeeds(t, db, 1)

	data := url.Values{}
	data.Add("difficulty", "1")
	data.Add("fun", "5")

	w := postHasStatus(t, db, "/vote/00-01-02-03-04", &data, http.StatusOK)
	body := w.Body.String()
	bodyHasFragments(t, body, []string{"Difficulty", "value=\"5\""})
}
