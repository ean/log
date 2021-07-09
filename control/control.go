// Package control implements necessary functionality for runtime level
// toggling. A control file is used as level configuration. The
// file content is memory mapped in the process to do real time lookup
// when a log messages is created.
package control

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/juju/fslock"
	"ngrd.no/log/control/mmap"
)

var (
	// DefaultControlPath is where the default log control instance will store its control file
	DefaultControlPath string

	// logControl should only be access from MaybeNewGlobalLogControl
	logControl *LogControl
	l          sync.Mutex

	on  = levelValue{' ', ' ', 'O', 'N'}
	off = levelValue{' ', 'O', 'F', 'F'}

	// FATAL, ERROR, WARNING, INFO, DEBUG
	DefaultLevelString = fmt.Sprintf(" %3s %3s %3s %3s %3s", "ON", "ON", "ON", "ON", "OFF")
)

func init() {
	DefaultControlPath = os.Getenv("LOG_CONTROL_PATH")
	if DefaultControlPath == "" {
		DefaultControlPath = "/var/run/logcontrol"
	}
}

// MaybeNewGlobalLogControl returns the cached global LogControl instance.
// The LogControl instance is created on the first invocation of the function.
func MaybeNewGlobalLogControl() *LogControl {
	l.Lock()
	defer l.Unlock()
	if logControl == nil {
		logControl = NewLogControl(DefaultControlPath)
	}
	return logControl
}

// ApplicationAndComponentToKey builds the lookup key format used by S
func ApplicationAndComponentToKey(application, component string) string {
	return application + ":" + component
}

// Level is the type used to specify a log message level
// fatal, error, warning, info or debug
type Level int
type ControlPtr []byte
type WritableControlPtr ControlPtr
type levelValue []byte

func (p ControlPtr) ShouldLog(level Level) bool {
	offset := level - 1
	return bytes.Equal(p[offset*4:offset*4+4], on)
}

type LogControl struct {
	ControlPath     string
	controlLockPath string

	l       *sync.RWMutex
	memory  *mmap.MMap
	mapping map[string]*ControlLine
	fw      *os.File
}

func NewLogControl(controlPath string) *LogControl {
	return &LogControl{
		ControlPath:     controlPath,
		controlLockPath: controlPath + ".lock",
		l:               &sync.RWMutex{},
	}
}

func (c *LogControl) ReadControlFile() error {
	unlock, err := c.Lock()
	if err != nil {
		return err
	}
	defer unlock()
	return c.readControlFile()
}

func (c *LogControl) readControlFile() error {
	if c.memory == nil {
		m, err := mmap.Map(c.ControlPath, mmap.PROT_READ, mmap.MAP_SHARED)
		if err != nil {
			return fmt.Errorf("mmap map: %w", err)
		}
		c.memory = m
	} else {
		if err := c.memory.Extend(); err != nil {
			return fmt.Errorf("mmap extend: %w", err)
		}
	}
	c.mapping = map[string]*ControlLine{}

	if err := c.populateMapping(); err != nil {
		return fmt.Errorf("parse control file: %w", err)
	}
	return nil
}

func (c *LogControl) Lock() (func() error, error) {
	fl := fslock.New(c.controlLockPath)
	if err := fl.Lock(); err != nil {
		return nil, fmt.Errorf("fslock.New: %w", err)
	}
	c.l.Lock()
	return func() error {
		c.l.Unlock()
		return fl.Unlock()
	}, nil
}

func (c *LogControl) populateMapping() error {
	controlLines, err := parseControl(c.memory.Data)
	if err != nil {
		return err
	}
	for _, ctrl := range controlLines {
		c.mapping[ApplicationAndComponentToKey(ctrl.Application, ctrl.Component)] = ctrl
	}
	return nil
}

func (c *LogControl) Register(application, component string) error {
	unlock, err := c.Lock()
	if err != nil {
		return fmt.Errorf("lock: %w", err)
	}
	defer unlock()
	if c.memory == nil {
		err := c.readControlFile()
		if err != nil {
			return fmt.Errorf("read control file: %w", err)
		}
	}
	if err := c.memory.Extend(); err != nil {
		return err
	}
	if present := c.keyPresent(application, component); present {
		return nil
	}
	if c.fw == nil {
		f, err := os.OpenFile(c.ControlPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		c.fw = f
	}
	if len(c.memory.Data) == 0 {
		// add header
		// TODO: Correct log control file documentation link
		fmt.Fprintf(c.fw, "# log control file, modified by log-control\n")
		fmt.Fprintf(c.fw, "# See https://github.com/ean/log/blob/master/foo for details")
	}
	line := fmt.Sprintf("%s:%s%s\n", application, component, DefaultLevelString)
	c.fw.WriteString(line)
	pos, err := c.fw.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	// fmt.Printf("Position: %d\n", pos)
	err = c.memory.Extend()
	if err != nil {
		return fmt.Errorf("map: %w", err)
	}
	c.mapping[ApplicationAndComponentToKey(application, component)] = &ControlLine{
		Application: application,
		Component:   component,
		Ptr:         c.memory.Data[int(pos)-1-len(DefaultLevelString) : pos-1],
	}
	return nil
}

func (c *LogControl) ShouldLog(key string, level Level) bool {
	c.l.RLock()
	defer c.l.RUnlock()
	ptr := ControlPtr(DefaultLevelString)
	if cl, ok := c.mapping[key]; ok {
		ptr = cl.Ptr
	}
	return ptr.ShouldLog(level)
}
