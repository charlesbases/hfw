package webcode

import (
	"fmt"
	"strconv"

	"github.com/charlesbases/logger"
)

// ecodes register codes
var ecodes = map[int32]string{}

// add .
func add(e int32, m string) Code {
	if _, found := ecodes[e]; found {
		logger.Fatalf("ecode: %d already existed.", e)
	}
	ecodes[e] = m
	return Parse(e)
}

// Code is an int error code spec
type Code int32

// Int32 .
func (e Code) Int32() int32 {
	return int32(e)
}

// Error .
func (e Code) Error() string {
	if m, found := ecodes[e.Int32()]; found {
		return m
	}
	return fmt.Sprintf("unknown ecode: %d", e)
}

// Parse parse code int to error
func Parse(e int32) Code {
	return Code(e)
}

// String parse code string to error
func String(e string) Code {
	if len(e) == 0 {
		return StatusOK
	}
	if e, err := strconv.Atoi(e); err != nil {
		return ServerErr
	} else {
		return Code(e)
	}
}
