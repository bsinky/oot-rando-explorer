package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/bsinky/sohrando/authentication"
	"github.com/gin-gonic/gin"
)

func TestLoginAuth(t *testing.T) {
	t.Parallel()
	app := FreshDb(t)

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	SetupRouter(router, &app.App)

	LogInUser(t, app.TestUser2, w, ctx, router)
}

func TestAdminPanelUserOps(t *testing.T) {
	db := FreshDb(t)

	// create user
	_, err := authentication.CreateUser(db.App.DB, "foo", "initial")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// make admin
	if err := authentication.SetAdmin(db.App.DB, "foo", true); err != nil {
		t.Fatalf("SetAdmin: %v", err)
	}
	u, _ := authentication.GetUser(db.App.DB, "foo")
	if u == nil || !u.IsAdmin {
		t.Fatalf("expected foo to be admin")
	}

	// make moderator
	if err := authentication.SetModerator(db.App.DB, "foo", true); err != nil {
		t.Fatalf("SetModerator: %v", err)
	}
	u, _ = authentication.GetUser(db.App.DB, "foo")
	if u == nil || !u.IsModerator {
		t.Fatalf("expected foo to be moderator")
	}

	// reset password and validate
	if err := authentication.ResetPassword(db.App.DB, "foo", "newpass"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	u, _ = authentication.GetUser(db.App.DB, "foo")
	if match, _ := u.PasswordMatches("newpass"); !match {
		t.Fatalf("password did not match after reset")
	}
	if match, _ := u.PasswordMatches("initial"); match {
		t.Fatalf("old password still valid after reset")
	}
}

func TestOperationsOnMissingUser(t *testing.T) {
	db := FreshDb(t)

	if err := authentication.SetAdmin(db.App.DB, "nope", true); err == nil {
		t.Fatalf("expected error when setting admin on missing user")
	}
	if err := authentication.SetModerator(db.App.DB, "nope", true); err == nil {
		t.Fatalf("expected error when setting moderator on missing user")
	}
	if err := authentication.ResetPassword(db.App.DB, "nope", "pw"); err == nil {
		t.Fatalf("expected error when resetting password on missing user")
	}
}
