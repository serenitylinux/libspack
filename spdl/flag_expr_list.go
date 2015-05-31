package spdl

// Flagexprlist represents a " " seperated list of flags
type Flagexprlist []FlagExpr

func (fsl Flagexprlist) String() string {
	res := ""
	for _, fs := range fsl {
		res += fs.String() + " "
	}
	return res
}

//Returns the Flags with default states
func (fsl Flagexprlist) Defaults() FlatFlagList {
	res := make(FlatFlagList, 0)
	for _, fs := range fsl {
		res[fs.Flag.Name] = fs.Flag
	}
	return res
}

func (fsl Flagexprlist) Contains(name string) bool {
	for _, fl := range fsl {
		if fl.Flag.Name == name {
			return true
		}
	}
	return false
}

func (fsl Flagexprlist) Verify(list FlatFlagList) bool {
	for _, fs := range fsl {
		if !fs.Verify(list) {
			return false
		}
	}
	return true
}
