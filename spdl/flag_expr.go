package spdl

import (
	"errors"
	"strings"

	"github.com/serenitylinux/libspack/parser"
)

type FlagExpr struct {
	Flag FlatFlag
	list *exprlist
}

func fromString(s string) (fs FlagExpr, err error) {
	s = strings.Replace(s, " ", "", -1)
	in := parser.NewInput(s)

	f, err := ParseFlat(&in)
	if err != nil {
		return fs, err
	}
	fs.Flag = f

	if exists := in.HasNext(1); !exists {
		//No conditions for flag
		return
	}

	if s, _ := in.Next(1); s != "(" {
		return fs, errors.New("Missing '(' after flag")
	}

	var l *exprlist
	l, err = parseexprlist(&in)
	if err != nil {
		return
	}
	fs.list = l

	if s, _ := in.Next(1); s != ")" {
		err = errors.New("Missing ')' at the end of input")
		return
	}

	if exists := in.HasNext(1); exists {
		err = errors.New("Trailing chars after end of flag definition: '" + in.Rest() + "'")
		return
	}
	return
}

func (f FlagExpr) Verify(list FlatFlagList) bool {
	if list.IsEnabled(f.Flag.Name) {
		return f.list.Enabled(list)
	}

	return true
}

func (f FlagExpr) String() string {
	return f.Flag.ColorString() + "(" + f.list.String() + ")"
}
