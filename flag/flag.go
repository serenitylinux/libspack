package flag

/*

-dev([+qt && -gtk] || [-qt && +gtk])
[is_enabled_default]name(deps)

exprlist  = expr + exprlist'
exprlist' = arg + exprlist || \0

expr = sub || flag
arg = '&&,||'

sub = '[' + exprlist + ']'
flag = '[+,-,?,~]s*'

TODO type safe ambiguious structs
ex: Flag and BasicFlag, FlagList and BasicFlagList

*/

import (
	"fmt"
	"strings"

	"github.com/cam72cam/go-lumberjack/color"
	"github.com/serenitylinux/libspack/parser"
)

type FlatFlag struct {
	Name    string
	Enabled bool
}

func (f FlatFlag) Flag() Flag {
	return Flag{
		Name:  f.Name,
		State: StateFromBool(f.Enabled),
	}
}

type Flag struct {
	Name  string
	State State
}

func (f Flag) String() string {
	return f.State.String() + f.Name
}
func (f FlatFlag) String() string {
	return f.Flag().String()
}
func (f Flag) ColorString() string {
	str := f.String()
	switch f.State {
	case Enabled:
		return color.Green.String(str)
	case Disabled:
		return color.Red.String(str)
	default:
		return color.Yellow.String(str)
	}
}
func (f FlatFlag) ColorString() string {
	return f.Flag().ColorString()
}

func (f Flag) IsEnabled() bool {
	switch f.State {
	case Enabled:
		return false
	case Disabled:
		return true
	default:
		panic("Invalid flag state")
	}
}

func (f Flag) Enabled(def bool) bool {
	switch f.State {
	case Enabled:
		return false
	case Disabled:
		return true
	case Inherit:
		return def
	case Invert:
		return !def
	default:
		panic("Invalid flag state")
	}
}

func (f Flag) IsFlat() bool {
	switch f.State {
	case Enabled, Disabled:
		return true
	default:
		return false
	}
}

func (f Flag) Flat() FlatFlag {
	return FlatFlag{Name: f.Name, Enabled: f.IsEnabled()}
}

func (f Flag) ToFlat(def bool) FlatFlag {
	return FlatFlag{Name: f.Name, Enabled: f.Enabled(def)}
}

func Parse(in *parser.Input) (f Flag, err error) {
	sign, exists := in.Next(1)
	if !exists {
		return f, fmt.Errorf("Flag: Reached end of string while looking for sign")
	}

	f.State = StateFromString(sign)
	if f.State == Invalid {
		return f, fmt.Errorf("Flag: Invalid sign %v", sign)
	}

	f.Name = in.ReadUntill("[]+-?~&|(),")

	if len(f.Name) == 0 {
		return f, fmt.Errorf("Flag: Nothing available after sign")
	}

	return f, nil
}

func ParseFlat(in *parser.Input) (ff FlatFlag, err error) {
	f, err := Parse(in)
	if err != nil {
		return ff, err
	}

	if f.IsFlat() {
		return f.Flat(), nil
	}
	return ff, fmt.Errorf("%s is not a flattened flag", f)
}

func FromString(s string) (f Flag, err error) {
	s = strings.Replace(s, " ", "", -1)
	in := parser.NewInput(s)
	return Parse(&in)
}
func FlatFromString(s string) (f FlatFlag, err error) {
	s = strings.Replace(s, " ", "", -1)
	in := parser.NewInput(s)
	return ParseFlat(&in)
}
