package domain

import "fmt"

var ErrInvalidUserCredentials = fmt.Errorf("invalid user credentials")

var ErrSecretAlreadyExists = fmt.Errorf("secret already exists")
var ErrInvalidSecretType = fmt.Errorf("invalid secret type")
var ErrSecretNotFound = fmt.Errorf("secret not found")
var ErrInvalidSecretID = fmt.Errorf("invalid secret id")
var ErrVersionConflict = fmt.Errorf("secret version conflict")

var ErrSaltNotSet = fmt.Errorf("kdf salt not set for user")
