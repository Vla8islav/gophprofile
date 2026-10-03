package domain

import (
	"errors"
	"fmt"
)

var ErrInvalidUserCredentials = fmt.Errorf("invalid user credentials")

var ErrAvatarNotFound = fmt.Errorf("avatar not found")
var ErrNotAvatarOwner = fmt.Errorf("you can only delete your own avatars")
var ErrUnsupportedAvatarFormat = fmt.Errorf("unsupported avatar format")
var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

// IsBusinessErr reports whether err is an expected business outcome
func IsBusinessErr(err error) bool {
	return errors.Is(err, ErrAvatarNotFound) ||
		errors.Is(err, ErrNotAvatarOwner) ||
		errors.Is(err, ErrInvalidUserCredentials) ||
		errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrUserAlreadyExists) ||
		errors.Is(err, ErrUnsupportedAvatarFormat)
}
