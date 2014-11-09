package depres

//Containing the Insanity to a single file (hopefully)
//Christian Mesh Feb 2014

//TODO check valid set of flags on a per package basis

import (
	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/dep"
	"github.com/serenitylinux/libspack/depres/pkgdep"
	"strings"
)

type DepResParams struct {
	IgnoreBDeps bool
}

//TODO log.Error.Println(all the things)
var indent int = 0

func DepTree(node *pkgdep.Node, params DepResParams) bool {
	indent++
	defer func() { indent-- }()

	debug := func(s string) {
		log.Debug.Format("%s%s %s", strings.Repeat("\t", indent), node.Control().UUID(), s)
	}
	debug("check")

	//We are already installed exact (checks version and flags as well)
	//And not a reinstall
	//And not being built
	if !node.ForgeOnly && node.IsInstalled() {
		debug("already installed " + node.PkgInfo().PrettyString() + " in " + node.Graph.DestDir)
		return true
	}

	//We do not need to be rechecked
	if !node.Dirty {
		debug("clean")
		return true
	}

	//We are being built and do not care about bdeps, I think we are done here
	if node.ForgeOnly && params.IgnoreBDeps {
		debug("Ignore bdeps")
		return true
	}

	node.Dirty = false //We will be making sure we are clean in the next step
	rethappy := true   //Clap your hands!

	var alldeps dep.DepList
	if node.ForgeOnly {
		alldeps = node.Control().ParsedBDeps()
	} else {
		alldeps = node.Control().ParsedDeps()
	}

	setflags := node.ComputedFlags()
	deps := alldeps.EnabledFromFlags(*setflags)

	//We are new or have been changed
	for _, dep := range deps {
		debug("Require: " + dep.String())

		depnode := node.Graph.Find(dep.Name)
		//We are not part of the graph yet
		if depnode == nil {
			depnode = node.Graph.Add(dep.Name, false)
			if depnode == nil {
				rethappy = false
				continue
			}
			if !depnode.ForgeOnly {
				depnode.AddRdepConstraints(strings.Repeat("\t", indent+1))
			}
		}

		if depnode.ForgeOnly {
			debug("too far down the rabbit hole: " + dep.Name)
			rethappy = false
			continue
		}

		//Will set to dirty if changed and add parent constraint
		if !depnode.AddParent(node, dep) {
			//We can't add this parent constraint
			debug("Cannot change " + dep.Name + " to " + dep.String())
			log.Error.Write([]byte("Conflicting package constraints on " + dep.Name + ":" + "\n"))
			depnode.Constraints.PrintError("\t")
			rethappy = false
			continue
		}

		//Continue down the rabbit hole ...
		if !DepTree(depnode, params) {
			debug("Not Happy " + dep.Name)
			rethappy = false
		}
	}

	// At this point we need to see if our original requirements have changed
	// There may be a scenario where A -> B -> ... -> A(+flag) !-> B
	// I guess we could create a phantom package B to force A(+flag)
	// even though B would never be installed...
	// Or we could just yell at the user if this scenario ever happened
	// and force tell them to set a flag to fix this inconsistency.
	// I like yelling at users, let's try that method first

	// We should probably check and see if our flags have changed at all through that last
	// dep loop.  If our deps are a super set of our deps before that last loop,
	// We should be fine, but will need to recurse once more.  If not, we have a serious
	// problem.  We should figure out what info would be best to print to
	// the user in this scenario...  meh, we can get to that later.

	aftersetflags := node.ComputedFlags()
	afterdeps := deps.EnabledFromFlags(*aftersetflags)

	if deps.IsSubset(afterdeps) {
		debug("Deps are ok!")
		if len(deps) != len(afterdeps) {
			//We need to recurse again
			node.Dirty = true
			rethappy = DepTree(node, params)
		}
	} else {
		debug("Deps are NOT ok")
		log.Error.Format("Could not resolve %s, conflicting circular dependency!", node.String())
		log.Info.Format("Before: %s", deps.String())
		log.Info.Format("After: %s", afterdeps.String())
		//TODO more debuging stuffs
		rethappy = false
	}

	debug("done")
	return rethappy
}

func FindToBuild(graph *pkgdep.Graph, params DepResParams, destdir string) (*pkgdep.Graph, bool) {
	log.Debug.Println("Finding packages to build:")

	orderedlist := pkgdep.NewGraph(destdir)
	visitedlist := pkgdep.NewGraph(destdir)

	happy := findToBuild(graph, orderedlist, visitedlist, params, destdir)
	visitedlist.Reverse() //See diagram below

	return visitedlist, happy
}

func findToBuild(graph, orderedtreelist, visitedtreelist *pkgdep.Graph, params DepResParams, destdir string) bool {
	indent++
	defer func() { indent-- }()

	debug := func(s string) {
		log.Debug.Format("%s %s", strings.Repeat("\t", indent), s)
	}

	//list of packages to build
	tobuild := pkgdep.NewGraph(graph.DestDir)

	//Find packages we have yet to build
	for _, node := range graph.Nodes {
		if visitedtreelist.Size() != 0 { //If we are passed the first layer, we don't care if the pacakge is latest
			node.IsLatest = false
		}
		if !node.ForgeOnly && (node.SpakgExists() || node.IsInstalled()) {
			debug("Have " + node.PkgInfo().PrettyString())
		} else {
			debug("Build " + node.PkgInfo().PrettyString())
			tobuild.Append(node)
		}
		continue

		if !node.SpakgExists() && !node.IsInstalled() || node.ForgeOnly {
			debug("Build " + node.PkgInfo().PrettyString())
			tobuild.Append(node)
		} else {
			debug("Have " + node.PkgInfo().PrettyString())
		}
	}

	happy := true //If you are happy and you know it clap your hands!!
	for _, node := range tobuild.Nodes {
		debug("To Build: " + node.PkgInfo().PrettyString())
		//We have not already been "built"
		if !visitedtreelist.Contains(node.Name) {
			//Create a new graph representing the build deps of node
			newroot := pkgdep.New(node.Name, node.Repo)
			newroot.Constraints = node.Constraints //This *should* be a deep copy

			newroot.Graph = pkgdep.NewGraph(graph.DestDir)

			//mark newroot read only
			newroot.ForgeOnly = true
			newroot.IsLatest = node.IsLatest

			//Add ourselves to existing builds
			visitedtreelist.Append(newroot)

			debug("Dep check " + node.PkgInfo().PrettyString())
			if !DepTree(newroot, params) {
				happy = false
				continue
			}

			debug("Find check " + node.PkgInfo().PrettyString())
			if !findToBuild(newroot.Graph, orderedtreelist, visitedtreelist, params, destdir) {
				happy = false
				log.Error.Println("From " + node.PkgInfo().PrettyString())
				break
				continue
			}

			//We now have our deps in a correct state so we can add ourselves to the order
			orderedtreelist.Append(newroot)
		} else {
			//We have been visited but are not satisfied yet !!! A -> (C, B), B -> (C, A) Invalid
			/*
				A visit
					C visit
						No deps
					C order
					B visit
						C visit and order    == OK
						A visited and !order == NOT OK
					B order
				A order

				existing in order signifies that the package is ok to go
			*/
			if !orderedtreelist.Contains(node.Name) {
				log.Error.Println("Build dep loop: " + node.PkgInfo().PrettyString())
				happy = false
			}
		}
		debug("Done " + node.PkgInfo().PrettyString())
	}
	return happy
}
