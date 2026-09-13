package broker

import (
	"errors"
)

// permanentError unfixable handler failure that redelivery doesn't fix
type permanentError struct{ err error }

func (e permanentError) Error() string { return "permanent: " + e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

// Permanent wraps err so the consumer knows retrying is pointless
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return permanentError{err: err}
}

// IsPermanent reports whether err (anywhere in its chain) was marked with
func IsPermanent(err error) bool {
	var p permanentError
	return errors.As(err, &p)
}
