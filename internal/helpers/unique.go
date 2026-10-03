package helpers

import (
	"fmt"
	"sync/atomic"
	"time"
)

var loginCounter atomic.Int64

func UniqueLogin(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), loginCounter.Add(1))
}
