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
	IsForge     bool
	IsReinstall bool
	IgnoreBDeps bool
	DestDir     string
}

//TODO log.Error.Println(all the things)
var indent int = 0

func DepTree(node *pkgdep.PkgDep, params DepResParams) bool {
	indent++
	defer func() { indent-- }()

	debug := func(s string) {
		log.Debug.Format("%s%s %s", strings.Repeat("\t", indent), node.Control().UUID(), s)
	}
	debug("check")

	//We are already installed exact (checks version and flags as well)
	//And not a reinstall
	//And not being built
	if !params.IsForge && node.IsInstalled(params.DestDir) && !params.IsReinstall {
		debug("already installed")
		return true
	}
	node.IsReinstall = params.IsReinstall

	//We do not need to be rechecked
	if !node.Dirty {
		debug("clean")
		return true
	}

	//We are being built and do not care about bdeps, I think we are done here
	if params.IsForge && params.IgnoreBDeps {
		debug("Ignore bdeps")
		return true
	}

	node.Dirty = false //We will be making sure we are clean in the next step
	rethappy := true   //Clap your hands!

	var alldeps dep.DepList
	if params.IsForge {
		alldeps = node.Control().ParsedBDeps()
	} else {
		alldeps = node.Control().ParsedDeps()
	}

	setflags := node.ComputedFlags()
	deps := alldeps.EnabledFromFlags(*setflags)

	//	isbdep := params.IsForge //Make a copy of isForge for later
	params.IsForge = false
	params.IsReinstall = false

	//We are new or have been changed
	for _, dep := range deps {
		debug("Require: " + dep.Name)

		depnode := node.Graph.Find(dep.Name)
		//We are not part of the graph yet
		if depnode == nil {
			depnode = node.Graph.Add(dep.Name, params.DestDir)
			if !depnode.ForgeOnly {
				depnode.AddRdepConstraints(params.DestDir, strings.Repeat("\t", indent+1))
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

func FindToBuild(graph *pkgdep.PkgDepList, params DepResParams) (*pkgdep.PkgDepList, bool) {
	log.Debug.Println("Finding packages to build:")

	orderedlist := make(pkgdep.PkgDepList, 0)
	visitedlist := make(pkgdep.PkgDepList, 0)

	happy := findToBuild(graph, &orderedlist, &visitedlist, params)
	visitedlist.Reverse() //See diagram below

	return &visitedlist, happy
}

func findToBuild(graph, orderedtreelist, visitedtreelist *pkgdep.PkgDepList, params DepResParams) bool {
	indent++
	defer func() { indent-- }()

	debug := func(s string) {
		log.Debug.Format("%s %s", strings.Repeat("\t", indent), s)
	}

	//list of packages to build
	tobuild := make(pkgdep.PkgDepList, 0)

	//Find packages we have yet to build
	for _, node := range *graph {
		//Not Build Only
		//Package exists exactly or is exactly installed or
		//Not first time through and has a goodenough installed
		if !node.ForgeOnly && (node.SpakgExists() || node.IsInstalled(params.DestDir) || len(*visitedtreelist) != 0 && node.AnyInstalled(params.DestDir)) {
			debug("Have " + node.PkgInfo().PrettyString())
		} else {
			debug("Build " + node.PkgInfo().PrettyString())
			tobuild.Append(node)
		}
		continue

		if !node.SpakgExists() && !node.IsInstalled(params.DestDir) && !node.AnyInstalled(params.DestDir) ||
			node.ForgeOnly {
			debug("Build " + node.PkgInfo().PrettyString())
			tobuild.Append(node)
		} else {
			debug("Have " + node.PkgInfo().PrettyString())
		}
	}

	happy := true //If you are happy and you know it clap your hands!!
	params.IsForge = true
	for _, node := range tobuild {
		debug("To Build: " + node.PkgInfo().PrettyString())
		//We have not already been "built"
		if !visitedtreelist.Contains(node.Name) {
			//Create a new graph representing the build deps of node
			newroot := pkgdep.New(node.Name, node.Repo)
			newroot.Constraints = node.Constraints //This *should* be a deep copy

			newrootgraph := make(pkgdep.PkgDepList, 0)
			newroot.Graph = &newrootgraph

			//mark newroot read only
			newroot.ForgeOnly = true

			//Add ourselves to existing builds
			visitedtreelist.Append(newroot)

			debug("Dep check " + node.PkgInfo().PrettyString())
			if !DepTree(newroot, params) {
				happy = false
				continue
			}

			if !findToBuild(&newrootgraph, orderedtreelist, visitedtreelist, params) {
				happy = false
				debug("dep failure")
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
				debug("Build dep loop")
				happy = false
			}
		}
		debug("Done " + node.PkgInfo().PrettyString())
	}
	return happy
}
