package pkgdep

import (
	"fmt"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack"
	"github.com/serenitylinux/libspack/misc"
)

type Graph struct {
	DestDir string
	Nodes   []*Node
}

func NewGraph(destdir string, nodes ...*Node) *Graph {
	return &Graph{DestDir: destdir, Nodes: nodes}
}

func (g *Graph) Contains(name string) bool {
	for _, item := range g.Nodes {
		if item.Name == name {
			return true
		}
	}
	return false
}

func (g *Graph) Append(e *Node) {
	g.Nodes = append(g.Nodes, e)
}

//http://blog.golang.org/slices Magics!
func (g *Graph) Prepend(e *Node) {
	g.Nodes = g.Nodes[0 : len(g.Nodes)+1] //Increase size by 1
	copy(g.Nodes[1:], g.Nodes[0:])        //shift array up by 1
	g.Nodes[0] = e                        //set new first element
}

func (g *Graph) Size() int {
	return len(g.Nodes)
}

func (g *Graph) Print() {
	i := 0
	for _, item := range g.Nodes {
		str := item.String() + " "
		i += len(str)
		if i > misc.GetWidth()-10 {
			fmt.Println()
			i = len(str)
		}
		fmt.Print(str)
	}
	fmt.Println()
}

//http://stackoverflow.com/a/19239850
func (g *Graph) Reverse() {
	for i, j := 0, len(g.Nodes)-1; i < j; i, j = i+1, j-1 {
		g.Nodes[i], g.Nodes[j] = g.Nodes[j], g.Nodes[i]
	}
}

func (g *Graph) Add(name string, latest bool) *Node {
	//Create new pkgdep node
	_, repo := libspack.GetPackageLatest(name)
	if repo == nil {
		log.Error.Println("Unable to find repo for ", name)
		return nil
	}

	node := New(name, repo)
	node.Graph = g
	g.Append(node)

	node.AddGlobalConstraints()

	if !node.Exists() {
		log.Error.Println(name, " unable to satisfy parents") //TODO more info
	}

	return node
}

func (g *Graph) ToInstall() *Graph {
	nodes := make([]*Node, 0)

	for _, pkg := range g.Nodes {
		if !pkg.ForgeOnly && !pkg.IsInstalled() {
			nodes = append(nodes, pkg)
		}
	}

	return NewGraph(g.DestDir, nodes...)
}
func (g *Graph) CheckPackageFlags() bool {
	for _, pkg := range g.Nodes {
		if !pkg.ValidFlags() {
			return false
		}
	}
	return true
}
func (g *Graph) Find(name string) *Node {
	for _, pkg := range g.Nodes {
		if pkg.Name == name {
			return pkg
		}
	}
	return nil
}
