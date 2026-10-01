// Package timeutil holds shared time helpers.
//
// Paisa stores every instant in UTC (TIMESTAMPTZ) and presents it in Indian
// Standard Time. India has no daylight saving, so a fixed +05:30 offset is
// exact and avoids a dependency on the system tz database.
package timeutil

import "time"

// IST is Indian Standard Time (UTC+05:30).
var IST = time.FixedZone("IST", 5*60*60+30*60)

// StartOfDayIST returns midnight (00:00 IST) of the day containing t.
func StartOfDayIST(t time.Time) time.Time {
	t = t.In(IST)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, IST)
}
