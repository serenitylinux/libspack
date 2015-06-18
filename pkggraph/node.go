package pkggraph

import (
	"fmt"

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

	changed bool

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

		changed: n.changed,

		control: n.control,
		pkginfo: n.pkginfo,
	}
}

func (n *Node) Changed() bool {
	return n.changed
}

func (n *Node) change() error {
	n.changed = true
	n.control = nil
	n.pkginfo = nil
	n.isInstalled = false
	n.isBin = false

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
		return n.control == nil || spdl.NewVersion(spdl.GT, n.control.Version).Accepts(c.Version)
	}

	switch n.Type {
	//try already installed
	case InstallConvenient:
		err := n.Repo.MapInstalledByName(n.Graph.root, n.Name, func(p repo.PkgInstallSet) {
			if matchControl(*p.Control) {
				//Check satisfies flags
				if flags.IsSubsetOf(p.PkgInfo.FlagStates) {
					n.control = p.Control
					n.pkginfo = p.PkgInfo
					n.isInstalled = true
				}
			}
		})
		if err != nil {
			return err
		}
		if n.pkginfo != nil {
			return nil
		}
		fallthrough
	//latest binary if available
	case InstallLatestBin:
		n.Repo.MapAvailableByName(n.Name, func(c control.Control, p pkginfo.PkgInfo) {
			if matchControl(c) {
				//Check satisfies flags
				if flags.IsSubsetOf(p.FlagStates) {
					n.control = &c
					n.pkginfo = &p
					n.isBin = true
				}
			}
		})
		if n.pkginfo != nil {
			return nil
		}
		fallthrough
	//latest src if available
	case InstallLatestSrc:
		n.Repo.MapTemplatesByName(n.Name, func(c control.Control) {
			if matchControl(c) {
				n.control = &c
			}
		})
		if n.control != nil {
			n.pkginfo = pkginfo.FromControl(n.control)
			return n.pkginfo.SetFlagStates(flags)
		}
		fallthrough
	default:
		//TODO better error
		return fmt.Errorf("Unable to find version of package %v", n.Name)
	}
}

func (n *Node) AddConstraint(dep spdl.Dep) error {
	n.rdeps.Add(Constraint{value: dep})
	return n.change()
}
func (n *Node) AddParentConstraint(parent string, dep spdl.Dep) error {
	if !n.rdeps.HasParent(parent) {
		n.rdeps.Add(Constraint{parent: &parent, value: dep})
		return n.change()
	}
	return nil
}
func (n *Node) RemoveParentConstraint(parent string) error {
	if n.rdeps.RemoveParent(parent) {
		return n.change()
	}
	return nil
}
func (n *Node) SetInstallType(typ InstallType) error {
	if n.Type < typ {
		return n.change()
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

func (n *Node) InInstalled() bool {
	return n.isInstalled
}
func (n *Node) HasBinary() bool {
	return n.isBin
}

func (n *Node) Hash() string {
	return fmt.Sprintf("%s::%s %v %d (%s)", n.Repo.Name, n.Name, n.changed, n.Type, n.rdeps.Hash(n.Graph))
}
