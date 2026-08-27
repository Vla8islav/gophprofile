package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileSink_WritesJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	sink := NewFileSink(path)
	t.Cleanup(func() { _ = sink.Close() })

	require.NoError(t, sink.Write(context.Background(),
		Event{Operation: "secret.create", UserID: 1, Status: 201}))
	require.NoError(t, sink.Write(context.Background(),
		Event{Operation: "secret.get", UserID: 1, SecretID: "abc", Status: 200}))
	require.NoError(t, sink.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	require.Len(t, lines, 2) // one JSON object per line

	var e0 Event
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &e0))
	require.Equal(t, "secret.create", e0.Operation)
	require.Equal(t, 201, e0.Status)

	var e1 Event
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &e1))
	require.Equal(t, "secret.get", e1.Operation)
	require.Equal(t, "abc", e1.SecretID)
	require.Equal(t, 200, e1.Status)
}
