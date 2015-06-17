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

	rdeps Constraints

	changed bool

	control *control.Control
	pkginfo *pkginfo.PkgInfo
}

func NewNode(name string, repo *repo.Repo, graph *Graph) *Node {
	return &Node{
		Name:  name,
		Repo:  repo,
		Graph: graph,
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

	versions, err := n.rdeps.VersionChecker(n.Graph)
	if err != nil {
		return err
	}
	flags, err := n.rdeps.Flags(n.Graph)
	if err != nil {
		return err
	}

	//TODO refactor GetPackageByVersionChecker
	n.control = n.Repo.GetPackageByVersionChecker(n.Name, versions)
	if n.control == nil { //Unable to find package control
		//TODO better error
		return fmt.Errorf("Unable to find version of package %v", n.Name)
	}

	n.pkginfo = pkginfo.FromControl(n.control)
	if err := n.pkginfo.SetFlagStates(flags); err != nil {
		return err
	}
	return nil
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

func (n *Node) Control() control.Control {
	return *n.control
}

func (n *Node) Pkginfo() pkginfo.PkgInfo {
	return *n.pkginfo
}

func (n *Node) IsEnabled() bool {
	return n.rdeps.AnyEnabled(n.Graph)
}
