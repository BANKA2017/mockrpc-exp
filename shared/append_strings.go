package shared

import "strings"

func AppendStrings(s ...string) string {
	return strings.Join(s, "")
}
