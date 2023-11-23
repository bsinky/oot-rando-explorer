package routes

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/bsinky/sohrando/randoseed"
	"github.com/gin-gonic/gin"
)

func testUploadSeed(t *testing.T, app *TestData, fileName string, w *httptest.ResponseRecorder, ctx *gin.Context, router *gin.Engine, status int, user *UserInfo) {
	t.Helper()

	routePath := "/uploadseed"

	var session *http.Cookie
	if user != nil {
		session = LogInUser(t, user, w, ctx, router)[0]
		// Create a new recorder to hold the Headers/Body of the /uploadseed request, otherwise it will only hold the login results
		*w = *httptest.NewRecorder()
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()
		part, err := writer.CreateFormFile("spoilerlog", fileName)
		if err != nil {
			t.Error(err)
		}

		file, err := os.Open("../test/" + fileName)
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
	if session != nil {
		req.AddCookie(session)
	}
	router.ServeHTTP(w, req)
	if status != w.Code {
		t.Fatalf("expected response code %d, got %d", status, w.Code)
	}
}

func TestSeedPage(t *testing.T) {
	t.Parallel()

	app := FreshDb(t)

	CreateSeeds(t, app, 3)

	w := GetHasStatus(t, app, "/s/00-01-02-03-04/", http.StatusOK)
	body := w.Body.String()
	BodyHasFragments(t, body, []string{randoseed.Versions[1], app.TestUser1.Username})
}

func TestUploadSeed(t *testing.T) {
	app := FreshDb(t)

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, &app.App)

	fileName := "04-94-01-69-66.json"
	testUploadSeed(t, app, fileName, w, ctx, router, http.StatusOK, app.TestUser2)

	uploadedFile, err := os.Open(path.Join(app.SpoilerSeedsDir, fileName))
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

	redirectHeader := w.Result().Header.Get("Hx-Location")
	if redirectHeader != "/s/04-94-01-69-66" {
		t.Fatalf("Upload should have redirected to created seed page; got redirect header: %s", redirectHeader)
	}
}

func TestUploadRequiresLogIn(t *testing.T) {
	app := FreshDb(t)

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, &app.App)

	fileName := "04-94-01-69-66.json"
	// passing nil user to try uploading a seed without signing in first
	testUploadSeed(t, app, fileName, w, ctx, router, http.StatusUnauthorized, nil)

	uploadedFile, err := os.Open(path.Join(app.SpoilerSeedsDir, fileName))
	if !os.IsNotExist(err) {
		t.Fatal("File should not have been able to upload")
	}
	defer uploadedFile.Close()
}

func TestVoteOnSeed(t *testing.T) {
	t.Parallel()

	app := FreshDb(t)

	CreateSeeds(t, app, 1)

	data := url.Values{}
	data.Add("difficulty", "1")
	data.Add("fun", "5")

	w := PostHasStatus(t, app, "/s/00-01-02-03-04/vote", &data, http.StatusOK, app.TestUser2)
	body := w.Body.String()
	BodyHasFragments(t, body, []string{"Difficulty", "value=\"5\"", "Your rating"})
}
