package log

import (
	"io"
	"strconv"
	"time"

	"ngrd.no/log/control"
)

// Option sets option when a new logger is created
type Option func(l *Logger)

// WithTimeFormat sets how the logger should represent the time pf a log message
// using (t time.Time) Format(layout string).
func WithTimeLayout(layout string) Option {
	return func(l *Logger) {
		l.formatTime = func(t time.Time) string {
			return t.Format(layout)
		}
	}
}

// WithTimeUnixNano represents the time of a log message as nanoseconds
// since unix epoch.
func WithTimeUnixNano() Option {
	return func(l *Logger) {
		l.formatTime = func(t time.Time) string {
			return strconv.FormatInt(t.UnixNano(), 10)
		}
	}
}

// WithDisabledTimestamp removes the timestamp from the log message
func WithDisabledTimestamp() Option {
	return func(l *Logger) {
		l.formatTime = func(_ time.Time) string {
			return ""
		}
	}
}

// WithWriter sets a specific writer as the log message sink
func WithWriter(w io.Writer) Option {
	return func(l *Logger) {
		l.w = w
	}
}

// WithComponentName overrides the default component name for a Logger instance
func WithComponentName(component string) Option {
	return func(l *Logger) {
		l.component = component
	}
}

// WithLogControl overrides the default LogControl instance used
func WithLogControl(c *control.LogControl) Option {
	return func(l *Logger) {
		l.control = c
	}
}
