package spdl

import (
	"fmt"
	"strings"
)

type FlagList map[string]Flag
type FlatFlagList map[string]FlatFlag

func (l FlagList) String() string {
	res := make([]string, 0, len(l))
	for _, flag := range l {
		res = append(res, flag.String())
	}
	return strings.Join(res, " ")
}

func (l FlatFlagList) String() string {
	res := make([]string, 0, len(l))
	for _, flag := range l {
		res = append(res, flag.String())
	}
	return strings.Join(res, " ")
}

func (l FlagList) ColorString() string {
	res := make([]string, 0, len(l))
	for _, flag := range l {
		res = append(res, flag.ColorString())
	}
	return strings.Join(res, " ")
}

func (l FlatFlagList) ColorString() string {
	res := make([]string, 0, len(l))
	for _, flag := range l {
		res = append(res, flag.ColorString())
	}
	return strings.Join(res, " ")
}

func (l FlatFlagList) IsSubsetOf(ol FlatFlagList) bool {
	for _, flag := range l {
		if oflag, found := ol[flag.Name]; found {
			if oflag.Enabled != flag.Enabled {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (l FlatFlagList) IsEnabled(f string) bool {
	if flag, ok := l[f]; ok {
		return flag.Enabled
	}
	return false
}

func (l FlagList) Clone() FlagList {
	newl := make(FlagList, len(l))

	for i, flag := range l {
		newl[i] = Flag{flag.Name, flag.State}
	}

	return newl
}

func (l FlagList) WithDefaults(defaults FlatFlagList) (FlatFlagList, error) {
	newl := make(FlatFlagList)
	for _, flag := range l {
		if flag.IsFlat() {
			newl[flag.Name] = flag.Flat()
		} else {
			if def, ok := defaults[flag.Name]; ok {
				newl[flag.Name] = flag.FlatWithDefault(def.Enabled)
			} else {
				return nil, fmt.Errorf("Default for flag %s not found", flag.Name)
			}
		}
	}
	return newl, nil
}
