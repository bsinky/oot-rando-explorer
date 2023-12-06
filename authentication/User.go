package authentication

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserDisplay struct {
	ID          uint
	Username    string
	Avatar      string
	IsModerator bool
	IsAdmin     bool
}

func (u *UserDisplay) CanDeleteSeeds() bool {
	return u != nil && u.IsAdmin
}

type User struct {
	gorm.Model
	Username       string `gorm:"uniqueIndex"`
	HashedPassword string
	Avatar         string
	IsModerator    bool
	IsAdmin        bool
}

var (
	ErrUsernameAlreadyExists     = errors.New("this username is already in use")
	ErrUsernameOrPasswordInvalid = errors.New("username or password does not match")
)

func getRandomHashIconAvatar() string {
	hashIcon := rand.Intn(100) - 1
	return fmt.Sprintf("%02d", hashIcon)
}

func CreateUser(db *gorm.DB, username string, password string) (*User, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:       username,
		HashedPassword: hashedPassword,
		Avatar:         getRandomHashIconAvatar(),
	}

	if err = db.Save(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrUsernameAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

func GetUser(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func GetUserDisplayByID(db *gorm.DB, id uint) (*UserDisplay, error) {
	var user UserDisplay
	if err := db.Table("users").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (user *User) PasswordMatches(password string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, user.HashedPassword)
	if err != nil {
		return false, err
	}

	return match, nil
}

func GetCurrentUser(c *gin.Context) *UserDisplay {
	session := sessions.Default(c)
	maybeID := session.Get("User.ID")
	maybeUsername := session.Get("User.Username")
	maybeAvatar := session.Get("User.Avatar")
	maybeIsModerator := session.Get("User.IsModerator")
	maybeIsAdmin := session.Get("User.IsAdmin")
	if maybeUsername == nil || maybeAvatar == nil || maybeID == nil {
		return nil
	}

	isModerator, modOk := maybeIsModerator.(bool)
	isAdmin, admOk := maybeIsAdmin.(bool)

	return &UserDisplay{
		ID:          maybeID.(uint),
		Username:    maybeUsername.(string),
		Avatar:      maybeAvatar.(string),
		IsModerator: modOk && isModerator,
		IsAdmin:     admOk && isAdmin,
	}
}

func SetCurrentUser(c *gin.Context, user *User) error {
	session := sessions.Default(c)
	session.Set("User.ID", user.ID)
	session.Set("User.Username", user.Username)
	session.Set("User.Avatar", user.Avatar)
	session.Set("User.IsModerator", user.IsModerator)
	session.Set("User.IsAdmin", user.IsAdmin)
	return session.Save()
}

func LogoutCurrentUser(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}
