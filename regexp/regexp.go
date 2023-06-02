package regexp

import "regexp"

var (
	IP = regexp.MustCompile(`^(\d+\.){3}\d+$`)
)
