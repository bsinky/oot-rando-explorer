package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/randoseed"
	"github.com/bsinky/sohrando/util"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ViewWithUser interface {
	User() *authentication.UserDisplay
}

type LoginUserForm struct {
	Username string `binding:"required,max=30" form:"username"`
	Password string `binding:"required,max=100" form:"password"`
	Errors   []util.SimpleValidation
}

func (u *LoginUserForm) SetErrors(validationErr validator.ValidationErrors) {
	u.Errors = util.ToErrors(validationErr)
}

func (u *LoginUserForm) AddError(message string) {
	u.Errors = append(u.Errors, util.SimpleValidation{
		Message: message,
	})
}

type RegisterUserForm struct {
	LoginUserForm
	ConfirmPassword string `binding:"required,max=100,eqfield=Password" form:"confirmPassword"`
}

func AddUserRoutes(r *gin.Engine) {
	r.GET("/login", loginPage)
	r.GET("/logout", logoutAction)
	r.GET("/signup", signupPage)

	r.GET("/user/:id", userProfile)

	authGroup := r.Group("/user", authRequired())
	authGroup.GET("/avataroptions", avatarOptions)
	authGroup.POST("/changeavatar", userChangeAvatar)

	r.POST("/login/auth", loginGetAuthToken)
	r.POST("/signup/register", signupCreateUser)
}

func loginPage(c *gin.Context) {
	if authentication.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "login.html", nil)
}

func logoutAction(c *gin.Context) {
	if err := authentication.LogoutCurrentUser(c); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func userProfile(c *gin.Context) {
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)
	id := c.Param("id")

	if id == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	idValue, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var viewedUser *authentication.UserDisplay
	viewingOwnProfile := false
	if user != nil && user.ID == uint(idValue) {
		// signed in user viewing their own profile
		viewedUser = user
		viewingOwnProfile = true
	} else {
		viewedUser, err = authentication.GetUserDisplayByID(db, uint(idValue))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		} else if viewedUser == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	}

	uploadedSeeds, err := randoseed.UserUploadedSeeds(db, viewedUser.ID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "user.html", gin.H{
		"User":              user,
		"Viewed":            gin.H{"User": viewedUser},
		"ViewingOwnProfile": viewingOwnProfile,
		"UploadedSeeds":     uploadedSeeds,
	})
}

func avatarOptions(c *gin.Context) {
	// build list of possible avatars to choose from
	avatars := make([]string, 100)
	for i := range 100 {
		avatars[i] = fmt.Sprintf("%02d", i)
	}

	c.HTML(http.StatusOK, "avatarOptions.html", gin.H{
		"User":    authentication.GetCurrentUser(c),
		"Avatars": avatars,
	})
}

func userChangeAvatar(c *gin.Context) {
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// fetch the real User record
	var realUser authentication.User
	if err := db.First(&realUser, "id = ?", user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// read and validate the new avatar value from the form
	avatarVal := c.PostForm("avatar")
	if avatarVal == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	n, err := strconv.Atoi(avatarVal)
	if err != nil || n < 0 || n > 99 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	newAvatar := fmt.Sprintf("%02d", n)
	realUser.Avatar = newAvatar

	if err := db.Save(&realUser).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// update the session representation of the current user
	if err := authentication.SetCurrentUser(c, &realUser); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// redirect back to the user's profile
	util.HtmxRedirect(c, fmt.Sprintf("/user/%d", realUser.ID))
}

func loginGetAuthToken(c *gin.Context) {
	db := util.GetDatabase(c)

	userForm := &LoginUserForm{
		Errors: make([]util.SimpleValidation, 0),
	}

	renderErrors := func() {
		c.HTML(http.StatusOK, "loginform", &userForm)
	}

	if err := c.ShouldBind(userForm); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		userForm.SetErrors(validationErrors)
		renderErrors()
		return
	}

	user, err := authentication.GetUser(db, userForm.Username)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if user == nil {
		userForm.AddError(authentication.ErrUsernameOrPasswordInvalid.Error())
		renderErrors()
		return
	}

	ok, err := user.PasswordMatches(userForm.Password)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if !ok {
		userForm.AddError(authentication.ErrUsernameOrPasswordInvalid.Error())
		renderErrors()
		return
	}

	if err := authentication.SetCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Login successful, redirect back to main page
	util.HtmxRedirect(c, "/")
}

func signupPage(c *gin.Context) {
	if authentication.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "signup.html", nil)
}

func signupCreateUser(c *gin.Context) {
	db := util.GetDatabase(c)

	userForm := &RegisterUserForm{}

	renderErrors := func() {
		c.HTML(http.StatusOK, "signupform", &userForm)
	}

	if err := c.ShouldBind(userForm); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		userForm.SetErrors(validationErrors)
		renderErrors()
		return
	}

	user, err := authentication.CreateUser(db, userForm.Username, userForm.Password)
	if err != nil {
		userForm.AddError(err.Error())
		renderErrors()
		return
	}

	if err := authentication.SetCurrentUser(c, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Registration successful, redirect back to main page
	util.HtmxRedirect(c, "/")
}
