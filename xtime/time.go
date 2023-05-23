package xtime

import (
	"context"
	"time"
)

const (
	// DefaultLayout 默认时间格式
	DefaultLayout = "2006-01-02 15:04:05"
)

// Now 当前时间
func Now() string {
	return time.Now().Format(DefaultLayout)
}

// NowTimestamp 当前毫秒时间戳
func NowTimestamp() int64 {
	return time.Now().UnixMilli()
}

// Format 时间格式化输出
func Format(t time.Time) string {
	return t.Format(DefaultLayout)
}

// Timestamp 毫秒时间戳
func Timestamp(t time.Time) int64 {
	return t.UnixMilli()
}

// Parse 时间格式化字符串转时间
func Parse(s string) time.Time {
	t, _ := time.Parse(DefaultLayout, s)
	return t
}

// TimestampParse 时间格式化字符串转毫秒时间戳
func TimestampParse(s string) int64 {
	return Timestamp(Parse(s))
}

type Duration time.Duration

// Shrink will decrease the duration by comparing with context's timeout duration
// and return new timeout\context\CancelFunc.
func (d Duration) Shrink(c context.Context) (Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if ctimeout := time.Until(deadline); ctimeout < time.Duration(d) {
			// deliver small timeout
			return Duration(ctimeout), c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d))
	return d, ctx, cancel
}
