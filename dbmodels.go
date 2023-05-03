package uos

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

// AppUser represents a user account.
type AppUser struct {
	gorm.Model

	Name     string `gorm:"unique"`
	Language string

	PasswordHash string
	Salt         string

	IsAdmin bool

	csrfToken string
}

func (AppUser) TableName() string {
	return "internal_app_users"
}

// CreateAppUser creates, saves and returns a new AppUser object.
func CreateAppUser(name string, password string) (AppUser, error) {
	var (
		salt = passwordSalt()
		hash = passwordHash(password, salt)
	)

	if salt == nil || hash == "" {
		return AppUser{}, fmt.Errorf("could not create user: password hashing failed")
	}

	user := AppUser{
		Name: name,

		PasswordHash: hash,
		Salt:         base64encode(salt),
	}

	return user, DB.Create(&user).Error
}

// GetAppUser returns an AppUser object. Checks password.
func GetAppUser(name string, password string) (AppUser, error) {
	var user = AppUser{Name: name}
	err := DB.Where(&user).First(&user).Error
	if err != nil {
		return AppUser{}, err
	}

	hash := passwordHash(password, base64decode(user.Salt))
	if hash != user.PasswordHash {
		return AppUser{}, ErrorInvalidPassword
	}

	return user, nil
}

func passwordSalt() []byte {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		Log.WarnError("could not generate salt", err)
		return nil
	}

	return salt
}

func passwordHash(password string, salt []byte) string {
	hash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		Log.WarnError("could not generate password hash", err)
		return ""
	}

	return base64encode(hash)
}
