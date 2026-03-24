package timeutil

import (
	"testing"
	"time"
)

func TestParseHHMM(t *testing.T) {
	h, m, err := ParseHHMM("09:30")
	if err != nil {
		t.Fatalf("ParseHHMM() error = %v", err)
	}
	if h != 9 || m != 30 {
		t.Fatalf("expected 09:30, got %02d:%02d", h, m)
	}
}

func TestParseHHMM_Invalid(t *testing.T) {
	if _, _, err := ParseHHMM("bad"); err == nil {
		t.Fatal("expected error for invalid time")
	}
}

func TestWeekdayToISO(t *testing.T) {
	if got := WeekdayToISO(time.Monday); got != 1 {
		t.Fatalf("expected Monday=1, got %d", got)
	}
	if got := WeekdayToISO(time.Sunday); got != 7 {
		t.Fatalf("expected Sunday=7, got %d", got)
	}
}

func TestStartEndOfUTCDate(t *testing.T) {
	date := time.Date(2026, 3, 21, 15, 45, 0, 0, time.UTC)

	start, end := StartEndOfUTCDate(date)

	expectedStart := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	expectedEnd := expectedStart.Add(24 * time.Hour)

	if !start.Equal(expectedStart) {
		t.Fatalf("unexpected start: %v", start)
	}
	if !end.Equal(expectedEnd) {
		t.Fatalf("unexpected end: %v", end)
	}
}