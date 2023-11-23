package routes

import (
	"net/http/httptest"
	"testing"

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
