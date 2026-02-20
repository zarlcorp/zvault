package tui

import "time"

func fixedTime(year int, month time.Month, day int) func() time.Time {
	return func() time.Time {
		return time.Date(year, month, day, 12, 0, 0, 0, time.Local)
	}
}
