package util

// stupid SQL quoting but good enough for me.
func Quote(s string) string {
	if s == "" {
		return "NULL"
	}
	return "'" + s + "'"
}
