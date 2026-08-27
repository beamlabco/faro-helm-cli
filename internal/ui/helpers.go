package ui

import (
	"time"
)

// utcTimeToLocal converts a UTC time string (HH:MM:SS) to the user's local
// timezone, using today's date for DST-aware conversion. Returns "HH:MM:SS"
// in local time format. Falls back to the original string on parse errors.
func utcTimeToLocal(utcTime string) string {
	now := time.Now()
	today := now.Format("2006-01-02")

	parsed, err := time.Parse("2006-01-02 15:04:05", today+" "+utcTime)
	if err != nil {
		return utcTime
	}

	// parsed is in UTC; convert to local
	utcParsed := parsed.UTC()
	local := utcParsed.In(time.Local)
	return local.Format("15:04:05")
}

// getStatusIcon returns a short glyph for an attendance status.
func getStatusIcon(status string) string {
	switch status {
	case "present":
		return "✓"
	case "remote":
		return "🏠"
	case "half-day":
		return "½"
	case "absent":
		return "✗"
	default:
		return "•"
	}
}
