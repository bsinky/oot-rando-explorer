package randoseed

import "gorm.io/gorm"

// TODO: User is probably a good candidate for a separate package
type User struct {
	gorm.Model
	Username string
}

func GetUser(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
