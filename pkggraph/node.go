package pkggraph

import (
	"fmt"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/pkginfo"
	"github.com/serenitylinux/libspack/repo"
	"github.com/serenitylinux/libspack/spdl"
)

type Node struct {
	Name  string
	Repo  *repo.Repo
	Graph *Graph
	Type  InstallType

	rdeps Constraints

	isEnabled         bool
	inPath            bool
	hasNewConstraints bool

	control     *control.Control
	pkginfo     *pkginfo.PkgInfo
	isInstalled bool
	isBin       bool
}

func NewNode(name string, repo *repo.Repo, graph *Graph) *Node {
	return &Node{
		Name:  name,
		Repo:  repo,
		Graph: graph,
		Type:  InstallConvenient,
	}
}

func (n Node) Clone(newgraph *Graph) *Node {
	return &Node{
		Name:  n.Name,
		Repo:  n.Repo,
		Graph: newgraph,

		rdeps: n.rdeps.Clone(),

		isEnabled:         n.isEnabled,
		hasNewConstraints: n.hasNewConstraints,

		control:     n.control,
		pkginfo:     n.pkginfo,
		isInstalled: n.isInstalled,
		isBin:       n.isBin,
	}
}

func (n *Node) ApplyChanges() error {
	n.hasNewConstraints = false
	n.isInstalled = false
	n.isBin = false
	n.isEnabled = false

	log.Debug.Format(prefix+"Finding new version of %v", n.Name)

	var newControl *control.Control
	var newPkginfo *pkginfo.PkgInfo
	defer func() {
		n.control = newControl
		n.pkginfo = newPkginfo
		n.isEnabled = true
	}()

	versions, err := n.rdeps.Versions(n.Graph)
	if err != nil {
		return err
	}
	flags, err := n.rdeps.Flags(n.Graph)
	if err != nil {
		return err
	}

	matchControl := func(c control.Control) bool {
		//Check valid version
		for _, version := range versions {
			if !version.Accepts(c.Version) {
				return false
			}
		}
		//Check is latest
		return newControl == nil ||
			spdl.NewVersion(spdl.GT, newControl.Version).Accepts(c.Version) &&
				((newControl.Version == c.Version && newControl.Iteration < c.Iteration) || newControl.Version != c.Version)
	}

	switch n.Type {
	//try already installed
	case InstallConvenient:
		err := n.Repo.MapInstalledByName(n.Graph.root, n.Name, func(p repo.PkgInstallSet) {
			p.Control.FlagsCrunch()
			if matchControl(*p.Control) {
				//Check satisfies flags
				if flags.IsSubsetOf(p.PkgInfo.FlagStates) {
					newControl = p.Control
					newPkginfo = p.PkgInfo
					n.isInstalled = true
					n.isBin = true
				}
			}
		})
		if err != nil {
			return err
		}
		if newPkginfo != nil {
			return nil
		}
		fallthrough
	//latest binary if available
	case InstallLatestBin:
		n.Repo.MapAvailableByName(n.Name, func(c control.Control, p pkginfo.PkgInfo) {
			c.FlagsCrunch()
			if matchControl(c) {
				//Check satisfies flags
				if flags.IsSubsetOf(p.FlagStates) {
					newControl = &c
					newPkginfo = &p
					n.isBin = true
				}
			}
		})
		if newPkginfo != nil {
			return nil
		}
		fallthrough
	//latest src if available
	case InstallLatestSrc:
		n.Repo.MapTemplatesByName(n.Name, func(_ string, c control.Control) {
			c.FlagsCrunch()
			if matchControl(c) {
				newControl = &c
			}
		})
		if newControl != nil {
			newPkginfo = pkginfo.FromControl(newControl)
			return newPkginfo.SetFlagStates(flags)
		}
		fallthrough
	default:
		//TODO better error
		return fmt.Errorf("Unable to find version of package %v", n.Name)
	}
}

func (n *Node) AddConstraint(dep spdl.Dep) error {
	n.rdeps.Add(Constraint{value: dep})
	n.hasNewConstraints = true
	return nil
}
func (n *Node) AddParentConstraint(parent string, dep spdl.Dep) error {
	//TODO GLIBC HACK
	if n.Name == parent {
		return nil
	}

	if !n.rdeps.HasParent(parent) {
		log.Debug.Format(prefix+"Adding parent constraint %v by %v", dep.String(), parent)
		n.rdeps.Add(Constraint{parent: &parent, value: dep})
		n.hasNewConstraints = true
	}
	return nil
}
func (n *Node) RemoveParentConstraint(parent string) error {
	//TODO GLIBC HACK
	if n.Name == parent {
		return nil
	}

	if n.rdeps.RemoveParent(parent) {
		n.hasNewConstraints = true
	}
	return nil
}
func (n *Node) SetInstallType(typ InstallType) error {
	if n.Type < typ {
		return n.ApplyChanges()
	}
	return nil
}

func (n *Node) Control() control.Control {
	return *n.control
}

func (n *Node) Pkginfo() pkginfo.PkgInfo {
	return *n.pkginfo
}

func (n *Node) IsEnabled() bool {
	return n.rdeps.AnyEnabled(n.Graph)
}

func (n *Node) IsInstalled() bool {
	return n.isInstalled
}
func (n *Node) HasBinary() bool {
	return n.isBin
}

func (n *Node) Hash() string {
	return fmt.Sprintf("%s::%s %v %v %d (%s)", n.Repo.Name, n.Name, n.hasNewConstraints, n.isEnabled, n.Type, n.rdeps.Hash(n.Graph))
}
