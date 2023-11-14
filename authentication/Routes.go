package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type SimpleValidation struct {
	Message string
}

func (u *SimpleValidation) Error() string {
	return u.Message
}

func (u *SimpleValidation) FieldName() string {
	return ""
}

func AddRoutes(r *gin.Engine) {
	r.GET("/login", loginPage)
	r.GET("/logout", logoutAction)
	r.GET("/signup", signupPage)

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

	userForm := &LoginUserForm{
		Errors: make([]SimpleValidation, 0),
	}

	renderErrors := func() {
		c.HTML(http.StatusOK, "loginform", &userForm)
	}

	if err := c.Bind(userForm); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		userForm.SetErrors(validationErrors)
		renderErrors()
		return
	}

	user, err := GetUser(db, userForm.Username)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if user == nil {
		userForm.AddError(ErrUsernameOrPasswordInvalid.Error())
		renderErrors()
		return
	}

	ok, err := user.PasswordMatches(userForm.Password)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if !ok {
		userForm.AddError(ErrUsernameOrPasswordInvalid.Error())
		renderErrors()
		return
	}

	if err := SetCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Login successful, redirect back to main page
	// c.Redirect(http.StatusSeeOther, "/")
	c.Status(http.StatusOK)
	c.Header("HX-Location", "/")
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

	userForm := &RegisterUserForm{}

	renderErrors := func() {
		c.HTML(http.StatusOK, "signupform", &userForm)
	}

	if err := c.Bind(userForm); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		userForm.SetErrors(validationErrors)
		renderErrors()
		return
	}

	user, err := CreateUser(db, userForm)
	if err != nil {
		userForm.AddError(err.Error())
		renderErrors()
		return
	}

	if err := SetCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Registration successful, redirect back to main page
	c.Status(http.StatusOK)
	c.Header("HX-Location", "/")
}
