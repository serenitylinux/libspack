package spdl

type DepList []Dep

func (list *DepList) EnabledFromFlags(fs FlatFlagList) DepList {
	res := make(DepList, 0)
	for _, dep := range *list {
		//We have no include condition
		if dep.Condition == nil {
			res = append(res, dep)
			continue
		}

		if dep.Condition.Enabled(fs) {
			res = append(res, dep)
			break
		}
	}
	return res
}

func (list *DepList) Contains(dep Dep) bool {
	// I know this is awfull, need to impl a better equals at
	// point but I am quite lazy
	str := dep.String()
	for _, d := range *list {
		if d.String() == str {
			return true
		}
	}
	return false
}

func (list DepList) IsSubset(other DepList) bool {
	for _, dep := range list {
		if !other.Contains(dep) {
			return false
		}
	}
	return true
}

func (list *DepList) String() string {
	str := ""

	for _, d := range *list {
		str += d.String() + " "
	}

	return str
}
