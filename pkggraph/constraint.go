package pkggraph

import (
	"fmt"

	"github.com/serenitylinux/libspack/spdl"
)

//internals readonly
type Constraint struct {
	parent *string //Can't use Node here.  Clone becomes overly complex
	value  spdl.Dep
}

type Constraints []Constraint

func (c *Constraints) Add(val Constraint) {
	*c = append(*c, val)
}

func (c Constraints) HasParent(parent string) bool {
	for _, val := range c {
		if val.parent != nil && *val.parent == parent {
			return true
		}
	}
	return false
}

//https://github.com/golang/go/wiki/SliceTricks
func (c *Constraints) RemoveParent(parent string) bool {
	for i, val := range *c {
		if val.parent != nil && *val.parent == parent {
			(*c)[i], (*c)[len(*c)-1], *c = (*c)[len(*c)-1], Constraint{}, (*c)[:len(*c)-1]
			return true
		}
	}
	return false
}

func (c Constraints) Clone() Constraints {
	nc := make(Constraints, len(c))
	for i, val := range c {
		nc[i] = val //Don't need a deep clone, readonly value
	}
	return nc
}

func (c Constraints) Flags(g *Graph) (spdl.FlatFlagList, error) {
	var total spdl.FlatFlagList

	for _, val := range c {
		var parentFlags spdl.FlatFlagList //Will be empty if no parent

		if val.parent != nil {
			if parent, ok := g.nodes[*val.parent]; ok {
				parentFlags = parent.Pkginfo().FlagStates
			} else {
				return nil, fmt.Errorf("Invalid parent %v", *val.parent)
			}
		}

		if val.value.Condition != nil {
			if val.parent == nil {
				return nil, fmt.Errorf("Can't have a condition without a parent") //TODO better error
			}
			if !val.value.Condition.Enabled(parentFlags) {
				continue //Skip
			}
		}

		flags, err := val.value.Flags.WithDefaults(parentFlags)
		if err != nil {
			return nil, err
		}
		if err := total.Merge(flags); err != nil {
			return nil, err
		}

	}
	return total, nil
}

type VersionChecker func(string) bool

func (c Constraints) VersionChecker(g *Graph) (VersionChecker, error) {
	versions := make([]*spdl.Version, 0)

	for _, val := range c {
		var parentFlags spdl.FlatFlagList //Will be empty if no parent

		if val.parent != nil {
			if parent, ok := g.nodes[*val.parent]; ok {
				parentFlags = parent.Pkginfo().FlagStates
			} else {
				return nil, fmt.Errorf("Invalid parent %v", *val.parent)
			}
		}

		if val.value.Condition != nil {
			if val.parent == nil {
				return nil, fmt.Errorf("Can't have a condition without a parent") //TODO better error
			}
			if !val.value.Condition.Enabled(parentFlags) {
				continue //Skip
			}
		}

		if val.value.Version1 != nil {
			versions = append(versions, val.value.Version1)
		}
		if val.value.Version2 != nil {
			versions = append(versions, val.value.Version2)
		}
	}

	return func(str string) bool {
		for _, v := range versions {
			if !v.Accepts(str) {
				return false
			}
		}

		return true
	}, nil
}

func (c Constraints) AnyEnabled(g *Graph) bool {
	for _, val := range c {
		if val.value.Condition == nil {
			continue //OK
		}
		var flags spdl.FlatFlagList //Will be empty if no parent
		if val.parent != nil {
			if parent, ok := g.nodes[*val.parent]; ok {
				flags = parent.Pkginfo().FlagStates
			}
		}
		if val.value.Condition.Enabled(flags) {
			continue //OK
		}
		return false
	}
	return true
}
