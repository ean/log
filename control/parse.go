package control

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"
)

type WritableControlLine struct {
	*ControlLine
	Ptr WritableControlPtr
}

type ControlLine struct {
	Application string
	Component   string
	Ptr         ControlPtr
}

func (c *LogControl) keyPresent(application string, component string) bool {
	m := c.memory.Data
	str := ApplicationAndComponentToKey(application, component) + " "
	needle := []byte("\n" + str)
	return bytes.Contains(m, needle)
}

func parseControl(m []byte) ([]*ControlLine, error) {
	l := len(m)
	controlLines := []*ControlLine{}
	offset := 0
	for offset < l {
		start := offset
		i := offset
		for ; i < l; i++ {
			if m[i] == '\n' {
				break
			}
		}
		if m[start] == '#' {
			offset = i + 1
			continue
		}
		ctrl, err := parseControlLine(m[start:i])
		if err != nil {
			return nil, err
		}
		controlLines = append(controlLines, ctrl)
		offset = i + 1 // skip past newline
	}
	return controlLines, nil
}

func bytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{Data: bh.Data, Len: bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

func parseControlLine(line []byte) (*ControlLine, error) {

	colon := bytes.IndexByte(line, ':')
	if colon == -1 {
		return nil, fmt.Errorf("no application end mark found: %s", string(line))
	}
	if len(line) < colon+1 {
		return nil, fmt.Errorf("line too short to contain component and level string: %s", string(line))
	}
	space := bytes.IndexByte(line[colon+1:], ' ')
	if space == -1 {
		return nil, fmt.Errorf("no component end mark found: %s", string(line))
	}
	if len(line)-space < len(DefaultLevelString) {
		return nil, fmt.Errorf("full level toggle string not found: %s", string(line))
	}
	ptr := line[space : space+len(DefaultLevelString)]
	return &ControlLine{
		Application: bytesToString(line[0:colon]),
		Component:   bytesToString(line[colon+1 : space]),
		Ptr:         ptr,
	}, nil
}
