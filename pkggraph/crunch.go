package pkggraph

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

func (g *Graph) crunch(iters Iterations) error {
	//Save current graph
	hash := g.Hash()
	if _, ok := iters[hash]; ok {
		//TODO better debug
		return fmt.Errorf("Duplicate hash %v", hash)
	}
	iters[hash] = g.Clone()

	var indent int = 0
	var handle func(node *Node) error
	handle = func(node *Node) error {
		indent++
		defer func() { indent-- }()
		prefix := strings.Repeat("\t", indent)
		debug := func(s string) {
			log.Debug.Format(prefix + node.Pkginfo().String() + ":" + s)
		}

		//Has not changed since last iteration
		if !node.Changed() {
			//debug("Not changed")
			return nil
		}

		if !node.IsEnabled() {
			debug("Not needed")
			return nil
		}

		debug("Start")
		defer debug("Done")

		for _, dep := range node.Control().Deps {
			depnode, ok := g.nodes[dep.Name]
			if !ok {
				debug("Bad Dep: " + dep.Name)
				return fmt.Errorf("Dependency %v not found", dep.Name)
			}

			changed := depnode.Changed()

			depnode.AddParentConstraint(node.Name, dep)

			//Already hit during this iteration
			//This is as deep as we go
			if changed {
				debug(fmt.Sprintf("Already hit %v", dep.Name))
				continue
			}

			if err := handle(depnode); err != nil {
				return err
			}
		}
		return nil
	}

	//handle all
	log.Debug.Format("Handling nodes")
	for _, node := range g.nodes {
		if err := handle(node); err != nil {
			return err
		}
	}

	//prune changed nodes
	log.Debug.Format("Pruning nodes")
	for _, node := range g.nodes {
		if node.Changed() {
			node.changed = false
			log.Debug.Format("Pruning %v", node.Name)
			//Prune
			for _, n := range g.nodes {
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
		if node.Changed() {
			isDone = false
			break
		}
	}

	//Here we go again!
	if !isDone {
		log.Debug.Format("Iterating again!")
		return g.crunch(iters)
	}

	return nil
}
