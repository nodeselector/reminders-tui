package ui

import (
	"testing"
	"time"
)

func TestRelativeDate(t *testing.T) {
	now := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		due  time.Time
		want string
	}{
		{
			name: "today",
			due:  time.Date(2026, 3, 29, 6, 0, 0, 0, time.UTC),
			want: "today",
		},
		{
			name: "tomorrow",
			due:  time.Date(2026, 3, 30, 6, 0, 0, 0, time.UTC),
			want: "tomorrow",
		},
		{
			name: "yesterday",
			due:  time.Date(2026, 3, 28, 6, 0, 0, 0, time.UTC),
			want: "yesterday",
		},
		{
			name: "in 3 days",
			due:  time.Date(2026, 4, 1, 6, 0, 0, 0, time.UTC),
			want: "in 3 days",
		},
		{
			name: "5 days ago",
			due:  time.Date(2026, 3, 24, 6, 0, 0, 0, time.UTC),
			want: "5 days ago",
		},
		{
			name: "next week",
			due:  time.Date(2026, 4, 7, 6, 0, 0, 0, time.UTC),
			want: "next week",
		},
		{
			name: "far future same year",
			due:  time.Date(2026, 8, 15, 6, 0, 0, 0, time.UTC),
			want: "Sat Aug 15",
		},
		{
			name: "different year",
			due:  time.Date(2027, 1, 10, 6, 0, 0, 0, time.UTC),
			want: "Jan 10, 2027",
		},
		{
			name: "zero time",
			due:  time.Time{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RelativeDate(tt.due, now)
			if got != tt.want {
				t.Errorf("RelativeDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPriorityIndicator(t *testing.T) {
	// Just verify they don't panic and return non-empty
	tests := []string{"high", "medium", "low", "none", ""}
	for _, p := range tests {
		t.Run(p, func(t *testing.T) {
			got := PriorityIndicator(p)
			if got == "" {
				t.Errorf("PriorityIndicator(%q) returned empty", p)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"hello world this is long", 10, "hello w..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
