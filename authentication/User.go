package authentication

import (
	"fmt"
	"math/rand"

	"github.com/alexedwards/argon2id"
	"gorm.io/gorm"
)

type UserDisplay struct {
	ID       uint
	Username string
	Avatar   string
}

type User struct {
	gorm.Model
	Username       string
	HashedPassword string
	Avatar         string
}

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
		return nil, err
	}
	return user, nil
}

func GetUser(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.First(&user, "username = ?", username).Error; err != nil {
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
