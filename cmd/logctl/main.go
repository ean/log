package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"ngrd.no/log"
	"ngrd.no/log/control"
)

var aFlag = flag.String("a", "", "filter on application, default match all")
var cFlag = flag.String("c", "", "filter on component, default match all. If component ends with / it will match all components with specified prefix"+
	"§11  ")

type changeLevel struct {
	level control.Level
	on    bool
}

func (cl changeLevel) modify(p control.WritableControlPtr) {
	if cl.on {
		p.On(cl.level)
	} else {
		p.Off(cl.level)
	}
}

func main() {
	flag.Parse()

	c := control.NewLogControl(control.DefaultControlPath)
	update, err := c.OpenForUpdate()
	if err != nil {
		fmt.Printf("failed opening control file for update: %v\n", err)
		os.Exit(1)
	}

	lines, err := update.ParseControl()
	if err != nil {
		fmt.Printf("parsing control file: %v\n", err)
		os.Exit(1)
	}

	rest := flag.Args()
	changes := []changeLevel{}
	for _, x := range rest {
		changes = append(changes, parseChange(x)...)
	}

	for _, l := range filter(*aFlag, *cFlag, lines) {
		for _, c := range changes {
			c.modify(l.Ptr)
		}
		fmt.Printf("%s:%s%s\n", l.Application, l.Component, string(l.Ptr))
	}
	if len(changes) > 0 {
		if err := update.Flush(); err != nil {
			fmt.Printf("failed syncing data to file: %v", err)
			os.Exit(1)
		}
	}
}

func parseChange(x string) []changeLevel {
	if len(x) == 0 {
		fmt.Printf("change string can't be empty")
		os.Exit(1)
	}
	changes := []changeLevel{}
	on := true
	name := x
	switch x[0] {
	case '+':
		on = true
		name = x[1:]
	case '-':
		on = false
		name = x[1:]
	}
	if name == "all" {
		for _, l := range log.Levels {
			changes = append(changes, changeLevel{
				level: l,
				on:    on,
			})
		}
	} else {
		l := log.LevelStringToType(strings.ToUpper(name))
		if l == log.UNKNOWN {
			fmt.Printf("'%s' is not a known log level\n", name)
			os.Exit(1)
		}
		changes = []changeLevel{
			{
				level: l,
				on:    on,
			},
		}
	}
	return changes
}

func filter(application string, component string, lines []*control.WritableControlLine) []*control.WritableControlLine {
	filtered := []*control.WritableControlLine{}
	for _, l := range lines {
		if application != "" {
			if l.Application != application {
				continue
			}
		}
		if component != "" {
			if strings.HasSuffix(component, "/") {
				if !strings.HasPrefix(l.Component, component) {
					continue
				}
			} else if l.Component != component {
				continue
			}
		}
		filtered = append(filtered, l)
	}
	return filtered
}
