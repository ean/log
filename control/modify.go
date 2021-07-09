package control

import (
	"fmt"

	"ngrd.no/log/control/mmap"
)

type LogControlForUpdate struct {
	*LogControl
	memory *mmap.MMap
}

func (c *LogControl) OpenForUpdate() (*LogControlForUpdate, error) {
	unlock, err := c.Lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	m, err := mmap.Map(c.ControlPath, mmap.PROT_WRITE|mmap.PROT_READ, mmap.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("control file: %w", err)
	}
	lc := &LogControlForUpdate{
		LogControl: c,
		memory:     m,
	}
	return lc, nil
}

func (c *LogControlForUpdate) ParseControl() ([]*WritableControlLine, error) {
	cl, err := parseControl(c.memory.Data)
	if err != nil {
		return nil, err
	}
	wcl := make([]*WritableControlLine, len(cl))
	for i, cl := range cl {
		wcl[i] = &WritableControlLine{
			ControlLine: cl,
			Ptr:         WritableControlPtr(cl.Ptr),
		}
	}
	return wcl, nil
}

func (c *LogControlForUpdate) Flush() error {
	return c.memory.Flush()
}

func (c *LogControlForUpdate) Close() error {
	return c.memory.Unmap()
}

// On enables a log level
func (p WritableControlPtr) On(level Level) {
	for i, x := range on {
		p[int(level-1)*4+i] = x
	}
}

// Off disables a log level
func (p WritableControlPtr) Off(level Level) {
	for i, x := range off {
		p[int(level-1)*4+i] = x
	}
}

// ShouldLog is a proxy for (p ControlPtr) ShouldLog(Level)
func (p WritableControlPtr) ShouldLog(level Level) bool {
	return ControlPtr(p).ShouldLog(level)
}
