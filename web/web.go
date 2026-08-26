// Package web holds the embedded web interface form yandex example
package web

import "embed"

//go:embed static
var Static embed.FS
