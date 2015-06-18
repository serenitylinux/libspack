package pkggraph

import (
	"fmt"

	"github.com/serenitylinux/libspack/repo"
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

func NewGraph(root string, repos ...*repo.Repo) (*Graph, error) {
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

func (g *Graph) EnablePackage(dep spdl.Dep) bool {
	curr, ok := g.nodes[dep.Name]
	if ok {
		curr.AddConstraint(dep)
	}
	return ok
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
	//TODO
	return "42"
}