package spdl

import (
	"errors"

	"github.com/serenitylinux/libspack/parser"
)

//Flag or list
type expr struct {
	list *ExprList
	flag FlatFlag
}

func parseExpr(in *parser.Input) (*expr, error) {
	e := new(expr)
	s, _ := in.Peek(1)
	switch s {
	case "(":
		in.Next(1)

		newl, err := parseExprList(in)
		if err != nil {
			return nil, err
		}

		if newl == nil {
			return nil, errors.New("[ ... ] must contain at least one flag")
		}

		e.list = newl

		s, _ := in.Next(1)
		if s != ")" {
			return nil, errors.New("Expression: Unexpected char '" + s + "'")
		}
	case "]", ")":
		//Done
	default:
		newf, err := ParseFlat(in)
		if err != nil {
			return nil, err
		}

		e.flag = newf
	}
	return e, nil
}
func (e *expr) verify(flist FlatFlagList) bool {
	if e.list != nil {
		return e.list.Enabled(flist)
	} else {
		return e.flag.Enabled == flist.IsEnabled(e.flag.Name)
	}
}
func (e *expr) String() string {
	if e.list != nil {
		return "(" + e.list.String() + ")"
	} else {
		return e.flag.String()
	}
}
