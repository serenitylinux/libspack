package spdl

// FlagExprList represents a " " seperated list of flags
type FlagExprList []FlagExpr

func (fsl FlagExprList) String() string {
	res := ""
	for _, fs := range fsl {
		res += fs.String() + " "
	}
	return res
}

//Returns the Flags with default states
func (fsl FlagExprList) Defaults() FlatFlagList {
	res := NewFlatFlagList(len(fsl))
	for _, fs := range fsl {
		res.Add(fs.Flag)
	}
	return res
}

func (fsl FlagExprList) Contains(name string) bool {
	for _, fl := range fsl {
		if fl.Flag.Name == name {
			return true
		}
	}
	return false
}

func (fsl FlagExprList) Verify(list FlatFlagList) bool {
	for _, fs := range fsl {
		if !fs.Verify(list) {
			return false
		}
	}
	return true
}
