package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joshu-sajeev/paisa/internal/timeutil"
)

func TestParseDateRangeDefaultsToISTMonth(t *testing.T) {
	// This instant is still September in IST although it is October in UTC.
	now := time.Date(2026, 9, 30, 18, 15, 0, 0, time.UTC)
	start, end, err := parseDateRange(httptest.NewRequest("GET", "/jars", nil), now)
	if err != nil {
		t.Fatalf("parseDateRange() error = %v", err)
	}
	wantStart := time.Date(2026, 9, 1, 0, 0, 0, 0, timeutil.IST)
	wantEnd := time.Date(2026, 10, 1, 0, 0, 0, 0, timeutil.IST)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("parseDateRange() = [%s, %s), want [%s, %s)", start, end, wantStart, wantEnd)
	}
}
