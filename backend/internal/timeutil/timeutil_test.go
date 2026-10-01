package timeutil_test

import (
	"testing"
	"time"

	"github.com/joshu-sajeev/paisa/internal/timeutil"
)

func TestStartOfDayIST(t *testing.T) {
	// 20:00 UTC on 2 Jan is 01:30 IST on 3 Jan.
	in := time.Date(2026, 1, 2, 20, 0, 0, 0, time.UTC)
	got := timeutil.StartOfDayIST(in)

	want := time.Date(2026, 1, 3, 0, 0, 0, 0, timeutil.IST)
	if !got.Equal(want) {
		t.Fatalf("StartOfDayIST() = %v, want %v", got, want)
	}
}
