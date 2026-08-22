package utils

import "time"

// TimestampFormats lists the order we try when parsing a timestamp that may
// come from either the SQLite driver (which returns "2006-01-02 15:04:05" via
// datetime('now')) or the pgx/PostgreSQL driver (which returns RFC3339 /
// ISO-8601). Keeping the list here in one place means the repository layer
// stays portable across both backends.
var TimestampFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.000",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05.000",
}

// ParseTimestamp tries the portable timestamp formats in order and returns the
// first match, or zero-time on failure. Prefer this over calling time.Parse
// directly with a single hardcoded format in repository code.
func ParseTimestamp(s string) (time.Time, error) {
	s = trim(s)
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range TimestampFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}

func trim(s string) string {
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}