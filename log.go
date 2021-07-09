// Package log implements logging package with runtime support for enabling and
// disabling log levels.
package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"ngrd.no/log/control"
)

var (
	ApplicationName string
	GlobalOptions   []Option
)

func init() {
	ApplicationName = filepath.Base(os.Args[0])
}

type Logger struct {
	control    *control.LogControl
	component  string
	key        string
	w          io.Writer
	formatTime func(t time.Time) string
}

func New(options ...Option) (*Logger, error) {
	l := &Logger{}

	caller := getFrame(1)
	l.component = getComponentFromCaller(caller)

	// Default options on logger
	WithWriter(os.Stdout)(l)
	WithTimeLayout(time.RFC3339)(l)

	allOptions := append(GlobalOptions, options...)

	// Apply options to look for LogControl override
	for _, option := range allOptions {
		option(l)
	}
	if l.control == nil {
		l.control = control.MaybeNewGlobalLogControl()
	}

	l.key = control.ApplicationAndComponentToKey(ApplicationName, l.component)
	if err := l.control.Register(ApplicationName, l.component); err != nil {
		return nil, fmt.Errorf("registering logger to log control failed: %w", err)
	}
	return l, nil
}

func (l *Logger) Log(level control.Level, msg string) {
	if l.control.ShouldLog(l.key, level) {
		s := len(msg)
		if s > 0 && msg[s-1] == '\n' {
			msg = msg[:s-1]
		}
		t := time.Now()
		l.w.Write([]byte(l.formatTime(t)))
		l.w.Write([]byte{'\t'})
		l.w.Write([]byte(l.component))
		l.w.Write([]byte{'\t'})
		l.w.Write([]byte(levelMap[level]))
		l.w.Write([]byte{'\t'})
		l.w.Write([]byte(msg))
		l.w.Write([]byte{'\n'})
	}
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Print(args...)
	os.Exit(1)
}

func (l *Logger) Fatalln(args ...interface{}) {
	l.Println(args...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Printf(format, args...)
	os.Exit(1)
}

func (l *Logger) Print(args ...interface{}) {
	l.Log(INFO, fmt.Sprint(args...))
}

func (l *Logger) Println(args ...interface{}) {
	l.Log(INFO, fmt.Sprintln(args...))
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.Log(INFO, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log(ERROR, fmt.Sprintf(format, args...))
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Log(WARNING, fmt.Sprintf(format, args...))
}
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(INFO, fmt.Sprintf(format, args...))
}
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(DEBUG, fmt.Sprintf(format, args...))
}
