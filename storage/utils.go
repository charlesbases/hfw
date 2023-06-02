package storage

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/charlesbases/hfw/regexp"
)

// CheckBucketName .
func CheckBucketName(v string) error {
	if len(strings.TrimSpace(v)) == 0 {
		return errors.New("bucket name cannot be empty")
	}
	if regexp.IP.MatchString(v) {
		return errors.New("bucket name cannot be an ip address")
	}

	return nil
}

// CheckObjectName .
func CheckObjectName(v string) error {
	if len(strings.TrimSpace(v)) == 0 {
		return errors.New("object name cannot be empty")
	}
	if strings.HasSuffix(v, "/") {
		return errors.New("object name cannot end with '/'")
	}
	if !utf8.ValidString(v) {
		return errors.New("object name with non UTF-8 strings are not supported")
	}
	return nil
}
