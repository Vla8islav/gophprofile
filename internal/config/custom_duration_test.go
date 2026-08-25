package config

import (
	"testing"
	"time"
)

func TestCustomSecondsDuration_Set(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "parse standard duration seconds",
			input:   "5s",
			want:    5 * time.Second,
			wantErr: false,
		},
		{
			name:    "parse standard duration minutes",
			input:   "2m",
			want:    2 * time.Minute,
			wantErr: false,
		},
		{
			name:    "parse integer as seconds",
			input:   "10",
			want:    10 * time.Second,
			wantErr: false,
		},
		{
			name:    "parse negative integer as absolute seconds",
			input:   "-7",
			want:    7 * time.Second,
			wantErr: false,
		},
		{
			name:    "parse zero integer",
			input:   "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid value returns error",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid duration returns error",
			input:   "12xs",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var d OptionalSecondsDuration

			err := d.Set(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if d.Duration != tt.want {
				t.Fatalf("expected duration %v, got %v", tt.want, d.Duration)
			}
		})
	}
}

func TestCustomSecondsDuration_String(t *testing.T) {
	t.Parallel()
	d := OptionalSecondsDuration{Duration: 3 * time.Second}

	got := d.String()
	want := "3s"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestCustomSecondsDuration_UnmarshalText(t *testing.T) {
	t.Parallel()
	var d OptionalSecondsDuration

	err := d.UnmarshalText([]byte("15"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := 15 * time.Second
	if d.Duration != want {
		t.Fatalf("expected duration %v, got %v", want, d.Duration)
	}
}
