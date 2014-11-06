package pkgdep

import (
	"fmt"

	"github.com/cam72cam/go-lumberjack/color"
	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack"
	"github.com/serenitylinux/libspack/constraintconfig"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/dep"
	"github.com/serenitylinux/libspack/flag"
	"github.com/serenitylinux/libspack/pkginfo"
	"github.com/serenitylinux/libspack/repo"
)

/******************************************
Represents an installable package and it's rdeps
*******************************************/
type Node struct {
	Name        string
	Repo        *repo.Repo
	Dirty       bool
	IsReinstall bool
	ForgeOnly   bool
	IsLatest    bool

	Constraints ConstraintList

	Graph *Graph
}

func New(name string, r *repo.Repo) *Node {
	new_node := Node{Name: name, Repo: r, Dirty: true, ForgeOnly: false, IsLatest: false}
	new_node.Constraints = make(ConstraintList, 0)

	return &new_node
}
func (node *Node) String() string {
	return fmt.Sprintf("%s::%s(%s)", node.Repo.Name, node.Control().UUID(), node.ComputedFlags())
}

//note: old parents should be removed, so we should never need to modify an existing constraint
func (node *Node) AddParent(parent *Node, reason dep.Dep) bool {
	if !node.Constraints.Contains(parent) {
		node.Constraints.AppendParent(parent, reason)
		node.Dirty = true
	}
	return node.Exists()
}

func (node *Node) AddGlobalConstraints() {
	//Add global flags to new depnode
	globalconstraint, exists := constraintconfig.GetAll(node.Graph.DestDir)[node.Name]
	if exists {
		ok := true
		for _, f := range *globalconstraint.Flags {
			if !node.Control().ParsedFlags().Contains(f.Name) {
				log.Error.Format("Invalid flag %s, skipping", f.Name)
				ok = false
			}
		}
		if ok {
			node.Constraints.AppendOther("Global Package Config", globalconstraint)
		}
	}
}

//Add rdeps if we are already installed
//We need to do this because changes to us might cause problems with packages that depend on us outside of the delta tree
//This links us into the existing "graph" on disk
func (node *Node) AddRdepConstraints(prefix string) {
	dep_info := node.Repo.GetInstalledByName(node.Name, node.Graph.DestDir)
	if dep_info != nil {
		for _, rdep := range libspack.RdepList(dep_info.PkgInfo) {
			//Copy pasta from depres
			depnode := node.Graph.Find(rdep.Control.Name)
			//We are not part of the graph yet
			if depnode == nil {
				log.Debug.Println(prefix + "Adding rdep:" + rdep.Control.Name)
				depnode = node.Graph.Add(rdep.Control.Name, false)
				depnode.AddRdepConstraints(prefix)
			}

			var reason dep.Dep
			found := false

			curr_control := depnode.Control()
			curr_pkginfo := depnode.PkgInfo()

			all_flags := curr_pkginfo.ComputedFlagStates() //rdep.PkgInfo.ComputedFlagStates()
			all_deps := curr_control.ParsedDeps()          //rdep.Control.ParsedDeps()

			for _, d := range all_deps.EnabledFromFlags(all_flags) {
				if d.Name == node.Name {
					reason = d
					found = true
					break
				}
			}

			//This really should not happen
			if !found {
				log.Error.Printlnf("Unable to figure out why %s is a dep of %s, we may have reconstructed the on disk graph of packages incorrectly",
					node.Name, depnode.Name)
				break
			}

			//Should return true, if not we have a serious problem with the packages on disk
			if !node.AddParent(depnode, reason) {
				log.Error.Write([]byte("Conflicting package constraints on " + color.Red.String("already installed") + " package " + node.Name + ":" + "\n"))
				depnode.Constraints.PrintError("\t")
			}
		}
	}
}

func (node *Node) Exists() bool {
	return node.Control() != nil && node.ComputedFlags() != nil
}

func (node *Node) Installed() *repo.PkgInstallSet {
	return node.Repo.GetInstalledByName(node.Name, node.Graph.DestDir)
}

func (node *Node) Control() *control.Control {
	checker := node.Constraints.ComputedVersionChecker()
	if !node.IsLatest { //Check if a version exists and is within the version constraints
		if installed := node.Installed(); installed != nil && checker(installed.Control.Version) {
			return installed.Control
		}
	}
	return node.Repo.GetPackageByVersionChecker(node.Name, checker)
}

func (node *Node) PkgInfo() *pkginfo.PkgInfo {
	flags := node.ComputedFlags()
	if flags == nil {
		return nil
	}

	checker := node.Constraints.ComputedVersionChecker()

	if !node.IsLatest { //Try to use the currently installed package
		installed := node.Installed()
		if installed != nil && installed.PkgInfo.Satisfies(*flags) && checker(installed.PkgInfo.Version) {
			return installed.PkgInfo
		}
	}

	//Not using installed, build from scratch
	p := pkginfo.FromControl(node.Control())
	err := p.SetFlagStates(*flags)
	if err != nil {
		log.Error.Println(err)
		return nil
	}
	return p
}

func (node *Node) ComputedFlags() *flag.FlagList {
	return node.Constraints.ComputedFlags(node)
}

func (node *Node) ValidFlags() bool {
	flagexpr := node.Control().ParsedFlags()
	flagstates := node.ComputedFlags()
	return flagexpr.Verify(flagstates)
}

func (node *Node) SpakgExists() bool {
	return node.Repo.HasSpakg(node.PkgInfo())
}

func (node *Node) IsInstalled() bool {
	pkginfo := node.PkgInfo()
	if node.IsReinstall || pkginfo == nil {
		return false
	}

	if node.IsLatest {
		return node.Repo.IsInstalled(pkginfo, node.Graph.DestDir)
	} else {
		return node.Repo.AnyInstalled(node.Name, node.Constraints.PkgsOnly().Deps(), node.Graph.DestDir)
	}
}
