package helpers

import (
	"fmt"
	"time"
)

func UniqueLogin(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
