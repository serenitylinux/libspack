package pkggraph

import (
	"crypto/md5"
	"fmt"
	"io"

	"github.com/serenitylinux/libspack"
	"github.com/serenitylinux/libspack/spdl"
)

/*
 Graph keeps track of all of the packages and their current states
 The internal order is kept consistant in ordered
 A quick reference lookup by name is provided by idmap
 These will NOT change once NewGraph is called, although
 the value of the nodes will vary

 idmap is nessesary since golang randomizes the map iterator
*/
type Graph struct {
	root    string
	ordered []*Node
	nodes   map[string]*Node
}

func NewGraph(root string, repos libspack.RepoList) (*Graph, error) {
	g := &Graph{
		root:    root,
		ordered: make([]*Node, 0, 100),
		nodes:   make(map[string]*Node, 100),
	}

	for _, repo := range repos {
		for _, name := range repo.GetAllNames() {
			if curr, ok := g.nodes[name]; ok {
				return nil, fmt.Errorf("Duplicate package name %v::%v, %v::%v", repo.Name, name, curr.Repo.Name, curr.Name)
			}
			node := NewNode(name, repo, g)
			g.ordered = append(g.ordered, node)
			g.nodes[name] = node
		}
	}

	return g, nil
}

func (g *Graph) EnablePackage(dep spdl.Dep, typ InstallType) error {
	if curr, ok := g.nodes[dep.Name]; ok {
		if err := curr.SetInstallType(typ); err != nil {
			return err
		}
		return curr.AddConstraint(dep)
	}
	return fmt.Errorf("Unable to find package %v", dep.Name)
}

func (g *Graph) Find(name string) (*Node, bool) {
	node, ok := g.nodes[name]
	return node, ok
}

func (g Graph) Clone() *Graph {
	ng := &Graph{
		root:    g.root,
		ordered: make([]*Node, len(g.ordered)),
		nodes:   make(map[string]*Node, len(g.nodes)),
	}
	for i, n := range g.ordered {
		clone := n.Clone(ng)
		ng.ordered[i] = clone
		ng.nodes[n.Name] = clone
	}
	return ng
}

func (g Graph) Hash() string {
	h := md5.New()
	for _, node := range g.nodes {
		io.WriteString(h, node.Hash())
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
