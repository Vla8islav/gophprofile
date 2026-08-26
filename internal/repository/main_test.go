package repository

import (
	"context"
	"os"
	"testing"
)

// TestMain tears the shared container down explicitly, so cleanup does not
// depend on the Ryuk reaper (which some docker environments cannot run).
func TestMain(m *testing.M) {
	code := m.Run()
	if testPGContainer != nil {
		_ = testPGContainer.Terminate(context.Background())
	}
	os.Exit(code)
}
