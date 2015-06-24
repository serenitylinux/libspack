package crunch

import (
	"fmt"

	"github.com/cam72cam/go-lumberjack/log"
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
			log.Debug.Format(prefix+"Removing parent constraint %v: %v", parent, val.value.String())
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

func (c Constraints) Map(g *Graph, fn func(Constraint, spdl.FlatFlagList) error) error {
	for _, val := range c {
		var parentFlags spdl.FlatFlagList //Will be empty if no parent

		if val.parent != nil {
			if parent, ok := g.nodes[*val.parent]; ok {
				if parent.IsEnabled() {
					if parent.pkginfo == nil {
						panic("INVALID" + parent.Name)
					}
					parentFlags = parent.Pkginfo().FlagStates
				}
			} else {
				return fmt.Errorf("Invalid parent %v", *val.parent)
			}
		}

		if val.value.Condition != nil {
			if val.parent == nil {
				return fmt.Errorf("Can't have a condition without a parent") //TODO better error
			}
			if !val.value.Condition.Enabled(parentFlags) {
				continue //Skip, not enabled
			}
		}
		if err := fn(val, parentFlags); err != nil {
			return err
		}
	}
	return nil
}

func (c Constraints) Flags(g *Graph) (total spdl.FlatFlagList, err error) {
	total = spdl.NewFlatFlagList(len(c))
	err = c.Map(g, func(val Constraint, parentFlags spdl.FlatFlagList) error {
		if val.value.Flags != nil {
			flags, err := val.value.Flags.WithDefaults(parentFlags)
			if err != nil {
				return err
			}
			if err := total.Merge(flags); err != nil {
				for _, c := range c {
					var parent string
					if c.parent != nil {
						parent = g.nodes[*c.parent].Hash()
					}
					fmt.Printf("DEP: %s %s\n", parent, c.value.String())
				}
				return err
			}
		}
		return nil
	})
	return total, err
}

func (c Constraints) Versions(g *Graph) (versions []spdl.Version, err error) {
	err = c.Map(g, func(val Constraint, parentFlags spdl.FlatFlagList) error {
		if val.value.Version1 != nil {
			versions = append(versions, *val.value.Version1)
		}
		if val.value.Version2 != nil {
			versions = append(versions, *val.value.Version2)
		}
		return nil
	})
	return versions, err
}

func (c Constraints) AnyEnabled(g *Graph) bool {
	if len(c) == 0 {
		return false
	}
	for _, val := range c {
		if val.value.Condition == nil {
			return true
		}
		var flags spdl.FlatFlagList //Will be empty if no parent
		if val.parent != nil {
			if parent, ok := g.nodes[*val.parent]; ok {
				flags = parent.Pkginfo().FlagStates
			}
		}
		if val.value.Condition.Enabled(flags) {
			return true
		}
	}
	return false
}

//TODO Do we want a list of dep.Dep or will flags suffice?
func (c Constraints) Hash(g *Graph) string {
	flags, err := c.Flags(g)
	if err != nil {
		panic(err)
	}
	return flags.String()
}
