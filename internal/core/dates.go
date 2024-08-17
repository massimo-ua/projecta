package core

import "time"

func DateOrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}
