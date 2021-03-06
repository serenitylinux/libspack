package spdl

/*

[condition]        name      versionspec              (depends)
[+flag && -flag]  pkgname <>=version<>=version  (+flag -flag ?flag ?flag[+flag && -flag] ~flag)

PkgName:
all except "<>=("

Version:
>=version (multiple possible)
<=version (multiple possible)
==version (singular)

FlagSet:
(FlagStat,FlagStat, ...)
+name
-name
?name
~name

*/

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mcuadros/go-version"
	"github.com/serenitylinux/libspack/parser"
)

type Dep struct {
	Condition *ExprList
	Name      string
	Version1  *Version
	Version2  *Version
	Flags     *FlagList
}

func (d *Dep) String() string {
	var res string
	if d.Condition != nil {
		res = "[" + d.Condition.String() + "]"
	}
	res += d.Name + d.Version1.String() + d.Version2.String()
	if d.Flags != nil {
		res += "(" + d.Flags.String() + ")"
	}

	return res
}

const (
	GT = 1
	LT = 2
	EQ = 3
)

type Version struct {
	typ int
	ver string
}

func NewVersion(typ int, ver string) *Version {
	return &Version{typ, ver}
}

func (v *Version) String() string {
	s := ""
	if v == nil {
		return s
	}

	switch v.typ {
	case GT:
		s = ">="
	case LT:
		s = "<="
	case EQ:
		s = "=="
	}
	return s + v.ver
}
func (v *Version) Accepts(verstr string) bool {
	switch v.typ {
	case GT:
		return version.Compare(verstr, v.ver, ">=")
	case LT:
		return version.Compare(verstr, v.ver, "<=")
	case EQ:
		return version.Compare(verstr, v.ver, "==")
	}
	panic(errors.New(fmt.Sprintf("Invalid version value: %d", v.typ)))
}

func conditionPeek(in *parser.Input) bool {
	s, _ := in.Peek(1)
	return s == "["
}

func versionPeek(in *parser.Input) bool {
	s, _ := in.Peek(1)
	return s == ">" || s == "<" || s == "="
}

func ParseDep(s string) (Dep, error) {
	s = strings.Replace(s, " ", "", -1)
	in := parser.NewInput(s)
	var d Dep
	err := d.parse(&in)
	return d, err
}

func (d *Dep) parse(in *parser.Input) error {
	if conditionPeek(in) {
		in.Next(1)

		new, err := parseExprList(in)
		if err != nil {
			return err
		}

		d.Condition = new

		if !in.IsNext("]") {
			return errors.New("Expected ']' at end of condition")
		}
	}

	d.Name = in.ReadUntill("<>=()")
	if len(d.Name) == 0 {
		return errors.New("Must specify dep package name")
	}

	if versionPeek(in) {
		var new Version
		err := new.parse(in)
		if err != nil {
			return err
		}
		d.Version1 = &new
	}

	if versionPeek(in) && d.Version1.typ != EQ {
		var new Version
		err := new.parse(in)
		if err != nil {
			return err
		}
		d.Version2 = &new
	}

	//no requirements
	if !in.HasNext(1) {
		return nil
	}

	new := NewFlagList(0)
	err := parseFlagSet(&new, in)
	if err != nil {
		return err
	}
	d.Flags = &new

	if in.HasNext(1) {
		return errors.New("Finished parsing, trailing chars '" + in.Rest() + "'")
	}

	return nil
}

func parseFlagSet(s *FlagList, in *parser.Input) error {
	if !in.IsNext("(") {
		return errors.New("Expected '(' to start flag set")
	}

	for {
		flag, err := Parse(in)
		if err != nil {
			return err
		}

		//TODO maybe check if already exists
		s.Add(flag)

		str, _ := in.Peek(1)
		if str != "+" && str != "-" && str != "~" && str != "?" {
			//We are at the end
			in.Next(1)

			if str != ")" {
				return errors.New("Invalid char '" + str + "', expected ')'")
			}

			break
		}
	}
	return nil
}

func (v *Version) parse(in *parser.Input) error {
	s, _ := in.Next(2)
	switch s {
	case ">=":
		v.typ = GT
	case "<=":
		v.typ = LT
	case "==":
		v.typ = EQ
	default:
		return errors.New("Invalid condition '" + s + "', expected [<>=]=")
	}
	v.ver = in.ReadUntill("<>=(")
	if len(v.ver) == 0 {
		return errors.New("[<>=]= must be followed by a version")
	}
	return nil
}
