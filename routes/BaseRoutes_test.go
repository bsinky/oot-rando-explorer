package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/migration"
	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/randoseed/logic"
	"github.com/bsinky/sohrando/randoseed/scrubsanity"
	"github.com/bsinky/sohrando/randoseed/shopsanity"
	"github.com/bsinky/sohrando/randoseed/tokensanity"

	"github.com/gin-gonic/gin"
)

type UserInfo struct {
	ID       uint
	Username string
	Password string
}

type TestData struct {
	App
	TestUser1 *UserInfo
	TestUser2 *UserInfo
}

func createUser(t *testing.T, data *TestData, userForm *UserInfo) {
	t.Helper()

	if user, err := authentication.CreateUser(data.DB, userForm.Username, userForm.Password); err != nil {
		t.Fatalf("Error creating test user %s: %s", userForm.Username, err)
	} else if user == nil {
		t.Fatalf("Error creating test user %s: returned user was nil", userForm.Username)
	} else {
		userForm.ID = user.ID
	}
}

func FreshDb(t *testing.T, path ...string) *TestData {
	t.Helper()

	app := FreshDbWithoutMigrations(t, path...)
	if err := migration.MigrateDB(app.DB); err != nil {
		t.Fatalf("Failed to migrate db %s", err)
	}

	var (
		testUser1 = &UserInfo{
			Username: "test-user!ArbitraryName",
			Password: "test-password1",
		}
		testUser2 = &UserInfo{
			Username: "different-account?Yeah",
			Password: "password2",
		}
	)

	createUser(t, app, testUser1)
	app.TestUser1 = testUser1
	createUser(t, app, testUser2)
	app.TestUser2 = testUser2

	return app
}

func FreshDbWithoutMigrations(t *testing.T, path ...string) *TestData {
	t.Helper()

	var dbUri string

	// Note: path can be specified in an individual test for debugging
	// purposes -- so the db file can be inspected after the test runs.
	// Normally it should be left off so that a truly fresh memory db is
	// used every time.
	if len(path) == 0 {
		dbUri = ":memory:"
	} else {
		dbUri = path[0]
	}

	testData := &TestData{}

	app, err := SetUpDBAndStorage(dbUri)
	if err != nil {
		t.Fatalf("Error opening memory db: %s", err)
	}
	testData.App = *app
	return testData
}

func BodyHasFragments(t *testing.T, body string, fragments []string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected body to contain '%s', got %s", fragment, body)
		}
	}
}

func NotBodyHasFragments(t *testing.T, body string, fragments []string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected body to not contain '%s', got %s", fragment, body)
		}
	}
}

func LogInUser(t *testing.T, user *UserInfo, w *httptest.ResponseRecorder, ctx *gin.Context, router *gin.Engine) []*http.Cookie {
	t.Helper()

	form := url.Values{}
	form.Set("username", user.Username)
	form.Set("password", user.Password)
	req, err := http.NewRequestWithContext(ctx, "POST", "/login/auth", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("Error trying to sign in test user %s: %s", user.Username, err)
	}
	req.Header.Add("Content-Type", gin.MIMEPOSTForm)

	router.ServeHTTP(w, req)
	if http.StatusOK != w.Code {
		t.Fatalf("login: expected response code %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	NotBodyHasFragments(t, body, []string{authentication.ErrUsernameOrPasswordInvalid.Error()})

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("login: reported success but did not set a session cookie")
	}
	return cookies
}

// TODO: add args... allowing a user to be passed in, if one is passed then try to sign in as them before issueing the GET
func GetHasStatus(t *testing.T, app *TestData, path string, status int, logInAs ...*UserInfo) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, &app.App)

	if len(logInAs) > 0 {
		LogInUser(t, logInAs[0], w, ctx, router)
	}

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

func PostHasStatus(t *testing.T, app *TestData, path string, data *url.Values, status int, logInAs ...*UserInfo) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, &app.App)

	var session *http.Cookie
	if len(logInAs) > 0 {
		session = LogInUser(t, logInAs[0], w, ctx, router)[0]
	}

	req, err := http.NewRequestWithContext(ctx, "POST", path, strings.NewReader(data.Encode()))
	if err != nil {
		t.Errorf("got error: %s", err)
	}
	if session != nil {
		req.AddCookie(session)
	}
	router.ServeHTTP(w, req)
	if status != w.Code {
		t.Errorf("expected response code %d, got %d", status, w.Code)
	}
	return w
}

// Create test seeds, User uploaded by is TestUser1
// VersionID is set to the earliest Version
func CreateSeeds(t *testing.T, data *TestData, count int) []*randoseed.Seed {
	t.Helper()
	seeds := []*randoseed.Seed{}
	for i := 0; i < count; i++ {
		s := randoseed.Seed{
			VersionID:      uint(1),
			FileHash:       fmt.Sprintf("%02d-%02d-%02d-%02d-%02d", i, i+1, i+2, i+3, i+4),
			Logic:          logic.Glitchless,
			Shopsanity:     shopsanity.Off,
			Tokensanity:    tokensanity.Off,
			Scrubsanity:    scrubsanity.Off,
			UserIDUploader: data.TestUser1.ID,
			RawSettings:    &randoseed.RawSettings{SettingsJSON: "{}"},
		}
		if err := data.DB.Save(&s).Error; err != nil {
			t.Fatalf("error creating seed: %s", err)
		}
		seeds = append(seeds, &s)
	}
	return seeds
}

func TestEmptyDatabase(t *testing.T) {
	app := FreshDb(t)
	seeds, err := randoseed.MostRecent(app.DB, 10)
	if err != nil {
		t.Fatalf("Error querying recent seeds from fresh db: %s", err)
	}
	if len(seeds) != 0 {
		t.Errorf("Expected 0 seeds, got %d", len(seeds))
	}
}

func TestMainPage(t *testing.T) {
	t.Parallel()

	app := FreshDb(t)

	CreateSeeds(t, app, 2)

	w := GetHasStatus(t, app, "/", http.StatusOK)

	body := w.Body.String()

	if len(body) == 0 {
		t.Fatalf("expected response non-zero body length")
	}
	BodyHasFragments(t, body, []string{randoseed.Versions[1]})
}
