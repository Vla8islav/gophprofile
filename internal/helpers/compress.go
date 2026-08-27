// Package helpers helper functions that don't belong anywhere else
package helpers

import (
	"bytes"
	"compress/gzip"
)

func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	defer gz.Close()

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil { // important: flushes data
		return nil, err
	}

	return buf.Bytes(), nil
}
