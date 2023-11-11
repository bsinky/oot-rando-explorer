package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddRoutes(r *gin.Engine) {
	r.GET("/login", loginPage)
	r.GET("/logout", logoutAction)

	r.POST("/login/auth", loginGetAuthToken)
	r.POST("/signup/register", signupCreateUser)
}

func loginPage(c *gin.Context) {
	if GetCurrentUser(c) != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "login.html", nil)
}

func logoutAction(c *gin.Context) {
	if err := LogoutCurrentUser(c); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func loginGetAuthToken(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)

	userForm := &UserForm{}
	if err := c.Bind(userForm); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := GetUser(db, userForm.Username)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ok, err := user.PasswordMatches(userForm.Password)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	} else if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := SetCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Login successful, redirect back to main page
	c.Redirect(http.StatusSeeOther, "/")
}

func signupPage(c *gin.Context) {
	if GetCurrentUser(c) != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "signup.html", nil)
}

func signupCreateUser(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)

	userForm := &UserForm{}
	if err := c.Bind(userForm); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := CreateUser(db, userForm)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	} else if user == nil || user.ID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := SetCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Registration successful, redirect back to main page
	c.Redirect(http.StatusSeeOther, "/")
}
