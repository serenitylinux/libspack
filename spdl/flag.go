package spdl

/*

flag = '[+,-,?,~]s*'

*/

import (
	"errors"
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
	Expr  *ExprList
}

func (f Flag) String() string {
	var expr string
	if f.Expr != nil {
		expr = "(" + f.Expr.String() + ")"
	}
	return f.State.String() + f.Name + expr
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
		return true
	case Disabled:
		return false
	default:
		panic("Invalid flag state")
	}
}

func (f Flag) EnabledEval(ffl FlatFlagList) (bool, error) {
	switch f.State {
	case Enabled:
		return true, nil
	case Disabled:
		return false, nil
	case Inherit, Invert:
		var res bool
		if f.Expr == nil {
			if fl, ok := ffl.flags[f.Name]; ok {
				res = fl.Enabled
			} else {
				return false, errors.New("Dependent flag not found")
			}
		} else {
			res = f.Expr.Enabled(ffl) //TODO Expr Eval Error
		}
		if f.State == Invert {
			res = !res
		}
		return res, nil
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

func (f Flag) FlatWithDefault(ffl FlatFlagList) (FlatFlag, error) {
	en, err := f.EnabledEval(ffl)
	if err != nil {
		return FlatFlag{}, err
	}
	return FlatFlag{Name: f.Name, Enabled: en}, nil
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

	next, _ := in.Peek(1)
	if next == "(" && f.State != Enabled && f.State != Disabled {
		in.Next(1)
		l, err := parseExprList(in)
		if err != nil {
			return f, err
		}
		f.Expr = l

		if s, _ := in.Next(1); s != ")" {
			return f, errors.New("Missing ')' at the end of flag def")
		}
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
