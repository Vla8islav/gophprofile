package audit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnixTime_JSONRoundTrip(t *testing.T) {
	ut := UnixTime{Time: time.Unix(1_100_200_300, 0)}

	data, err := json.Marshal(ut)
	require.NoError(t, err)
	require.Equal(t, "1100200300", string(data)) // Unix seconds

	var back UnixTime
	require.NoError(t, json.Unmarshal(data, &back))
	require.Equal(t, int64(1_100_200_300), back.Unix())
}

func TestUnixTime_UnmarshalGarbage(t *testing.T) {
	var ut UnixTime
	require.Error(t, json.Unmarshal([]byte(`"not-a-number"`), &ut))
}
