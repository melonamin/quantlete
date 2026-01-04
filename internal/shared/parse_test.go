package shared

import (
	"testing"
	"time"
)

func TestParseDateParam(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantYMD [3]int // year, month, day
	}{
		{
			name:    "empty string",
			input:   "",
			wantOK:  false,
			wantYMD: [3]int{0, 0, 0},
		},
		{
			name:    "RFC3339 format",
			input:   "2024-03-15T10:30:00Z",
			wantOK:  true,
			wantYMD: [3]int{2024, 3, 15},
		},
		{
			name:    "RFC3339 with timezone offset",
			input:   "2024-03-15T10:30:00+02:00",
			wantOK:  true,
			wantYMD: [3]int{2024, 3, 15},
		},
		{
			name:    "simple date format",
			input:   "2024-03-15",
			wantOK:  true,
			wantYMD: [3]int{2024, 3, 15},
		},
		{
			name:    "invalid format",
			input:   "15/03/2024",
			wantOK:  false,
			wantYMD: [3]int{0, 0, 0},
		},
		{
			name:    "invalid date",
			input:   "not-a-date",
			wantOK:  false,
			wantYMD: [3]int{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseDateParam(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseDateParam(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && tt.wantOK {
				if got.Year() != tt.wantYMD[0] || int(got.Month()) != tt.wantYMD[1] || got.Day() != tt.wantYMD[2] {
					t.Errorf("ParseDateParam(%q) = %v, want year=%d month=%d day=%d",
						tt.input, got, tt.wantYMD[0], tt.wantYMD[1], tt.wantYMD[2])
				}
			}
			if !ok && !tt.wantOK {
				if !got.IsZero() {
					t.Errorf("ParseDateParam(%q) returned non-zero time for failed parse: %v", tt.input, got)
				}
			}
		})
	}
}

func TestParseDateParamPreservesTime(t *testing.T) {
	// RFC3339 should preserve the full timestamp including time portion
	input := "2024-03-15T14:30:45Z"
	got, ok := ParseDateParam(input)
	if !ok {
		t.Fatalf("ParseDateParam(%q) failed unexpectedly", input)
	}

	want, _ := time.Parse(time.RFC3339, input)
	if !got.Equal(want) {
		t.Errorf("ParseDateParam(%q) = %v, want %v", input, got, want)
	}
}
