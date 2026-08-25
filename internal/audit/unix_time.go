package audit

import (
	"strconv"
	"time"
)

type UnixTime struct {
	time.Time
}

// MarshalJSON uses a value receiver
func (t UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.Unix(), 10)), nil
}

func (t *UnixTime) UnmarshalJSON(data []byte) error {
	sec, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	t.Time = time.Unix(sec, 0)
	return nil
}
