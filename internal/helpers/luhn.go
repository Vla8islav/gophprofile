package helpers

import luhnmod10 "github.com/luhnmod10/go"

func ValidLuhn(number string) bool {
	return luhnmod10.Valid(number)
}
