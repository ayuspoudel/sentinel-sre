package cluster

import (
	"regexp"
	"strings"
)

var invalidDNSChars = regexp.MustCompile(`[^a-z0-9.-]`)

func ToDNS1123Name(s string) string {
	s = strings.ToLower(s)
	s = invalidDNSChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-.")
	return s
}
