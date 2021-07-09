package log

import "ngrd.no/log/control"

const (
	FATAL control.Level = iota + 1
	ERROR
	WARNING
	INFO
	DEBUG

	UNKNOWN = -1
)

var Levels = []control.Level{
	FATAL,
	ERROR,
	WARNING,
	INFO,
	DEBUG,
}

var levelMap = map[control.Level]string{
	FATAL:   "FATAL",
	ERROR:   "ERROR",
	WARNING: "WARN",
	INFO:    "INFO",
	DEBUG:   "DEBUG",
}

func LevelStringToType(level string) control.Level {
	for l, s := range levelMap {
		if s == level {
			return l
		}
	}
	if level == "WARNING" {
		return WARNING
	}
	return UNKNOWN
}
