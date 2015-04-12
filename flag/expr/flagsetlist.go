package expr

import "github.com/serenitylinux/libspack/flag"

// FlagSetList represents a " " seperated list of flags
type FlagSetList []FlagSet

func (fsl FlagSetList) String() string {
	res := ""
	for _, fs := range fsl {
		res += fs.String() + " "
	}
	return res
}

//Returns the Flags with default states
func (fsl FlagSetList) Defaults() flag.FlatFlagList {
	res := make(flag.FlatFlagList, 0)
	for _, fs := range fsl {
		res[fs.Flag.Name] = fs.Flag
	}
	return res
}

func (fsl FlagSetList) Contains(name string) bool {
	for _, fl := range fsl {
		if fl.Flag.Name == name {
			return true
		}
	}
	return false
}

func (fsl FlagSetList) Verify(list flag.FlatFlagList) bool {
	for _, fs := range fsl {
		if !fs.Verify(list) {
			return false
		}
	}
	return true
}
