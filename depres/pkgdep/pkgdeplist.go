package pkgdep

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack"
)

type Graph struct {
	DestDir string
	Nodes   []*Node
}

func NewGraph(destdir string, nodes ...*Node) *Graph {
	return &Graph{DestDir: destdir, Nodes: nodes}
}

func (g *Graph) AlphaOrdered() []*Node {
	res := make([]*Node, 0, g.Size())
	for _, item := range g.Nodes {
		set := false
		for i, curr := range res {
			if curr.Name > item.Name {
				res = res[0 : len(res)+1] //Increase size by 1
				copy(res[1+i:], res[i:])  //shift array up by X
				res[i] = item             //set new first element
				set = true
				break
			}
		}
		if !set {
			res = append(res, item)
		}
	}
	return res
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
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	for _, item := range g.AlphaOrdered() {
		var flagstr string
		flags := item.ComputedFlags()
		if len(*flags) > 0 {
			flagstr = fmt.Sprintf("(%s)", flags.String())
		}
		fmt.Fprintf(w, "\t%s\t%s\t%s\n", item.Repo.Name, item.Control().UUID(), flagstr)
	}
	w.Flush()
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
