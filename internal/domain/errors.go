package domain

import "fmt"

var ErrInvalidUserCredentials = fmt.Errorf("invalid user credentials")

var ErrAvatarNotFound = fmt.Errorf("avatar not found")
var ErrNotAvatarOwner = fmt.Errorf("you can only delete your own avatars")
var ErrUnsupportedAvatarFormat = fmt.Errorf("unsupported avatar format")
