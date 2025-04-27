package repository

import (
	"time"
)

func parseMySQLDate(dateStr string) time.Time {
	if dateStr == "0000-00-00 00:00:00" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
