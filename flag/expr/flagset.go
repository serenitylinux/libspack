package expr

import (
	"errors"
	"strings"

	"github.com/serenitylinux/libspack/flag"
	"github.com/serenitylinux/libspack/parser"
)

// Represents a flag state and a possible expression dependency
// +flag(-foo && +bar)
type FlagSet struct {
	Flag flag.FlatFlag
	list *ExprList
}

func (fs *FlagSet) UnmarshalJSON(data []byte) (err error) {
	*fs, err = fromString(string(data))
	return err
}

func (fs *FlagSet) MarshalJSON() ([]byte, error) {
	return []byte(fs.String()), nil
}

func fromString(s string) (fs FlagSet, err error) {
	s = strings.Replace(s, " ", "", -1)
	in := parser.NewInput(s)

	f, err := flag.ParseFlat(&in)
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

	var l *ExprList
	l, err = ParseExprList(&in)
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

func (f FlagSet) Verify(list flag.FlatFlagList) bool {
	if list.IsEnabled(f.Flag.Name) {
		return f.list.Enabled(list)
	}

	return true
}

func (f FlagSet) String() string {
	return f.Flag.ColorString() + "(" + f.list.String() + ")"
}
