package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

// FileSink writes audit events to a file as JSON lines
type FileSink struct {
	path             string
	filterDescriptor *os.File

	mu sync.Mutex
}

// NewFileSink creates a FileSink that writes audit events to path
func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

// Write appends e to the sink file as a JSON line
func (s *FileSink) Write(_ context.Context, e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if s.filterDescriptor == nil {
		s.filterDescriptor, err = os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}

	}
	_, err = s.filterDescriptor.Write(append(payload, '\n'))
	if err != nil {
		return err
	}
	return nil
}

// Close closes the underlying file. Safe to call more than once.
func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filterDescriptor == nil {
		return nil
	}
	err := s.filterDescriptor.Close()
	s.filterDescriptor = nil
	return err
}
