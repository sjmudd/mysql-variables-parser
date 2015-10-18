package sysvar

import (
	"fmt"
)

type Types map[string]string

type Info struct {
	name        string
	types       Types
	cmdline     Types
	scope       Types
	default_val Types
	dynamic     Types
}

func (i Info) LastSysvar() string {
	return i.name
}

func (i *Info) SaveName(name string) {
	i.name = name
}

func (i *Info) SaveCommandLine(cmd_line string) {
	if i.cmdline == nil {
		i.cmdline = make(Types)
	}
	if _, found := i.cmdline[i.name]; found && i.cmdline[cmd_line] != cmd_line {
		fmt.Println("WARNING: already found a sysvar type for:", i.name)
		fmt.Println("WARNING: current value:", i.cmdline[i.name])
		fmt.Println("WARNING: new value:", cmd_line)
	}
	i.cmdline[cmd_line] = cmd_line
}

// add to the name / type map.
// We'll find more than one value if there's a table with different values for different
// versions. If the type changes then we'll need to record the value for each version.
// Not sure if that's needed yet.
// If the type does not change then there's no problem.
// Otherwise issue a warning (error?)
func (i *Info) SaveType(name_type string) {
	if i.types == nil {
		i.types = make(Types)
	}
	if _, found := i.types[i.name]; found && i.types[name_type] != name_type {
		fmt.Println("WARNING: already found a sysvar type for:", i.name)
		fmt.Println("WARNING: current value:", i.types[i.name])
		fmt.Println("WARNING: new value:", name_type)
	}
	i.types[name_type] = name_type
}

func (i *Info) SaveScope(scope string) {
	if i.scope == nil {
		i.scope = make(Types)
	}
	if _, found := i.scope[i.name]; found && i.scope[i.name] != scope {
		fmt.Println("WARNING: already found a sysvar scope for:", i.name)
		fmt.Println("WARNING: current value:", i.scope[i.name])
		fmt.Println("WARNING: new value:", scope)
	}
	i.scope[i.name] = scope
}

// save the default_val settings
func (i *Info) SaveDefault(default_value string) {
	if i.default_val == nil {
		i.default_val = make(Types)
	}
	if _, found := i.default_val[i.name]; found && i.default_val[i.name] != default_value {
		fmt.Println("WARNING: already found a sysvar default_val for:", i.name)
		fmt.Println("WARNING: current value:", i.default_val[i.name])
		fmt.Println("WARNING: new value:", default_value)
	}
	i.scope[i.name] = default_value
}

// set dynamic
func (i *Info) SaveDynamic(dynamic string) {
	if i.dynamic == nil {
		i.dynamic = make(Types)
	}
	if _, found := i.dynamic[i.name]; found && i.dynamic[i.name] != dynamic {
		fmt.Println("WARNING: already found a sysvar dynamic for:", i.name)
		fmt.Println("WARNING: current value:", i.dynamic[i.name])
		fmt.Println("WARNING: new value:", dynamic)
	}
	i.dynamic[i.name] = dynamic
}

func (i *Info) Defaults() Types {
	return i.default_val
}
func (i *Info) Scopes() Types {
	return i.scope
}
func (i *Info) ColumnTypes() Types {
	return i.types
}
func (i *Info) CmdLines() Types {
	return i.cmdline
}

func (i *Info) Dynamics() Types {
	return i.dynamic
}
