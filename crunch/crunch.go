package crunch

import (
	"fmt"
	"strings"

	"github.com/cam72cam/go-lumberjack/log"
)

type Iterations map[string]*Graph

func (g *Graph) Crunch() error {
	iters := make(Iterations)
	return g.crunch(iters)
}

var prefix string

func (g *Graph) crunch(iters Iterations) error {
	//Save current graph
	hash := g.Hash()
	if _, ok := iters[hash]; ok {
		//TODO better debug
		return fmt.Errorf("Duplicate hash %v", hash)
	}
	iters[hash] = g.Clone()

	for _, node := range g.ordered {
		node.lastHash = node.rdeps.Hash(g)
	}

	var indent int = 0
	var handle func(node *Node) error
	handle = func(node *Node) error {
		indent++
		defer func() { indent--; prefix = strings.Repeat("\t", indent) }()
		prefix = strings.Repeat("\t", indent)
		debug := func(s string) {
			if node.pkginfo == nil {
				log.Debug.Format(prefix + node.Name + ":" + s)
			} else {
				log.Debug.Format(prefix + node.Pkginfo().PrettyString() + ":" + s)
			}
		}

		if !node.IsEnabled() {
			log.Debug.Format(prefix + node.Name + ":" + "Not Needed")
			return nil
		}

		if node.inPath == true {
			debug("Already visited")
			return nil
		}

		debug("Start")
		defer debug("Done")

		if node.control == nil { //initial setup
			if err := node.ApplyChanges(); err != nil {
				return err
			}
		}

		node.inPath = true
		defer func() { node.inPath = false }()

		for _, dep := range node.Control().Deps {
			depnode, ok := g.nodes[dep.Name]
			if !ok {
				debug("Bad Dep: " + dep.Name)
				return fmt.Errorf("Dependency %v not found", dep.Name)
			}

			depnode.AddParentConstraint(node.Name, dep)

			if err := handle(depnode); err != nil {
				return err
			}
		}
		return nil
	}

	//handle all
	log.Debug.Format("Handling nodes")
	for _, node := range g.ordered {
		if err := handle(node); err != nil {
			return err
		}
	}

	prefix = "\t"
	//prune changed nodes
	log.Debug.Format("Pruning nodes")
	for _, node := range g.ordered {
		if node.hasNewConstraints && node.lastHash != node.rdeps.Hash(g) {
			log.Debug.Format(prefix+"Pruning node %v", node.Name)
			if err := node.ApplyChanges(); err != nil {
				return err
			}
			//Prune
			for _, n := range g.ordered {
				n.RemoveParentConstraint(node.Name)
			}

			if err := handle(node); err != nil {
				return err
			}
		}
	}

	//check done or iterate again
	isDone := true
	for _, node := range g.nodes {
		if node.lastHash != node.rdeps.Hash(g) {
			isDone = false
			break
		}
	}

	//Here we go again!
	if !isDone {
		str := ""
		for _, pkg := range g.ToWield() {
			str += pkg.Pkginfo().String() + " "
		}
		log.Debug.Format("ToWIeld:" + str)
		log.Debug.Format("=======Iterating again!========")
		return g.crunch(iters)
	}

	return nil
}
