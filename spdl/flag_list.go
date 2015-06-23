package spdl

import (
	"fmt"
	"strings"
)

type FlagList struct {
	flags   map[string]Flag
	ordered []Flag
}
type FlatFlagList struct {
	flags   map[string]FlatFlag
	ordered []FlatFlag
}

func NewFlagList(capacity int) (f FlagList) {
	f.flags = make(map[string]Flag, capacity)
	f.ordered = make([]Flag, 0, capacity)
	return f
}
func NewFlatFlagList(capacity int) (f FlatFlagList) {
	f.flags = make(map[string]FlatFlag, capacity)
	f.ordered = make([]FlatFlag, 0, capacity)
	return f
}

func (l *FlagList) Add(f Flag) {
	for i, curr := range l.ordered {
		if curr.Name == f.Name {
			l.ordered[i] = f
			l.flags[f.Name] = f
			return
		}
	}
	l.flags[f.Name] = f
	l.ordered = append(l.ordered, f)
}
func (l *FlatFlagList) Add(f FlatFlag) {
	for i, curr := range l.ordered {
		if curr.Name == f.Name {
			l.ordered[i] = f
			l.flags[f.Name] = f
			return
		}
	}
	l.flags[f.Name] = f
	l.ordered = append(l.ordered, f)
}

func (l FlagList) Contains(name string) (Flag, bool) {
	f, ok := l.flags[name]
	return f, ok
}

func (l FlatFlagList) Contains(name string) (FlatFlag, bool) {
	f, ok := l.flags[name]
	return f, ok
}

func (l FlagList) Slice() []Flag {
	return l.ordered
}
func (l FlatFlagList) Slice() []FlatFlag {
	return l.ordered
}

func (l FlagList) String() string {
	res := make([]string, 0, len(l.ordered))
	for _, flag := range l.ordered {
		res = append(res, flag.String())
	}
	return strings.Join(res, " ")
}

func (l FlatFlagList) String() string {
	res := make([]string, 0, len(l.ordered))
	for _, flag := range l.ordered {
		res = append(res, flag.String())
	}
	return strings.Join(res, " ")
}

func (l FlagList) ColorString() string {
	res := make([]string, 0, len(l.ordered))
	for _, flag := range l.ordered {
		res = append(res, flag.ColorString())
	}
	return strings.Join(res, " ")
}

func (l FlatFlagList) ColorString() string {
	res := make([]string, 0, len(l.ordered))
	for _, flag := range l.ordered {
		res = append(res, flag.ColorString())
	}
	return strings.Join(res, " ")
}

func (l FlatFlagList) IsSubsetOf(ol FlatFlagList) bool {
	for _, flag := range l.ordered {
		if oflag, found := ol.flags[flag.Name]; found {
			if oflag.Enabled != flag.Enabled {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (l *FlatFlagList) Merge(o FlatFlagList) error {
	for _, of := range o.ordered {
		if lf, ok := l.flags[of.Name]; ok {
			if lf.Enabled != of.Enabled {
				return fmt.Errorf("Conflicting flags in merge %v", of.Name) //TODO better error
			}
		} else {
			l.Add(of)
		}
	}
	return nil
}

func (l FlatFlagList) IsEnabled(f string) bool {
	if flag, ok := l.flags[f]; ok {
		return flag.Enabled
	}
	return false
}

func (l FlagList) Clone() FlagList {
	newl := NewFlagList(len(l.ordered))

	for _, flag := range l.ordered {
		newl.Add(Flag{flag.Name, flag.State, flag.Expr})
	}

	return newl
}

func (l FlagList) WithDefaults(defaults FlatFlagList) (FlatFlagList, error) {
	newl := NewFlatFlagList(len(l.ordered))
	for _, flag := range l.ordered {
		flat, err := flag.FlatWithDefault(defaults)
		if err != nil {
			return newl, err
		}
		newl.Add(flat)
	}
	return newl, nil
}

func (l FlatFlagList) ToFlagList() FlagList {
	newl := NewFlagList(len(l.ordered))
	for _, flat := range l.ordered {
		newl.Add(flat.Flag())
	}
	return newl
}
