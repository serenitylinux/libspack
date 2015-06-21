package libspack

import (
	"fmt"

	"github.com/cam72cam/go-lumberjack/color"
	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/pkggraph"
	"github.com/serenitylinux/libspack/pkginfo"
	"github.com/serenitylinux/libspack/repo"
	"github.com/serenitylinux/libspack/spdl"
)

//TODO: reinstall

func Forge(pkgs []spdl.Dep, root string, ignoreBDeps bool) error {
	return buildGraphs(pkgs, true, root, ignoreBDeps, false, pkggraph.InstallConvenient)
}

func Wield(pkgs []spdl.Dep, root string, reinstall bool, itype pkggraph.InstallType) error {
	return buildGraphs(pkgs, false, root, false, reinstall, itype)
}

func buildGraphs(pkgs []spdl.Dep, isForge bool, root string, ignoreBDeps bool, reinstall bool, itype pkggraph.InstallType) error {
	type forgeInfo struct {
		Graph    *pkggraph.Graph
		Pkginfo  *pkginfo.PkgInfo
		Control  *control.Control
		Template string
	}

	graph, err := pkggraph.NewGraph(root, repo.GetAllRepos())
	if err != nil {
		return fmt.Errorf("Unable to load package graph: %v", err.Error())
	}

	var addToForge func(spdl.Dep) error
	var addToWield func([]spdl.Dep, *pkggraph.Graph, pkggraph.InstallType) error

	toForge := make(map[string]*forgeInfo, 0)
	added := make(map[string]bool)
	forgeOrder := make([]string, 0)

	addToForge = func(pkg spdl.Dep) error {
		log.Info.Format("Forge %v", pkg.String())
		info := &forgeInfo{
			Graph: graph.Clone(),
		}

		r, err := repo.GetRepoFor(pkg.Name)
		if err != nil {
			return err
		}
		r.MapTemplatesByName(pkg.Name, func(t string, c control.Control) {
			if info.Control == nil || spdl.NewVersion(spdl.GT, info.Control.Version).Accepts(c.Version) {
				info.Control = &c
				info.Template = t
			}
		})

		if info.Control == nil {
			return fmt.Errorf("Unable to find package %v", pkg.Name)
		}

		flags, err := pkg.Flags.WithDefaults(nil)
		if err != nil {
			return err
		}

		info.Pkginfo = pkginfo.FromControl(info.Control)
		err = info.Pkginfo.SetFlagStates(flags)
		if err != nil {
			return err
		}

		key := info.Pkginfo.String()

		//Check if we have already or are in the process of adding
		if done, ok := added[key]; ok && done {
			return nil //Already done
		} else if ok && !done {
			return fmt.Errorf("Can not forge %v, it requires itself (%v) to build", pkg.Name, info.Pkginfo.PrettyString())
		}

		added[key] = false

		bdeps := make([]spdl.Dep, 0)
		for _, bdep := range info.Control.Bdeps {
			if bdep.Condition != nil && !bdep.Condition.Enabled(info.Pkginfo.FlagStates) {
				continue //Not enabled
			}
			defaults := make(spdl.FlatFlagList)
			if bdep.Flags != nil {
				defaults, err = bdep.Flags.WithDefaults(info.Pkginfo.FlagStates)
				if err != nil {
					return err
				}
			}

			flags := defaults.ToFlagList()
			bdeps = append(bdeps, spdl.Dep{
				Name:     bdep.Name,
				Version1: bdep.Version1,
				Version2: bdep.Version2,
				Flags:    &flags,
			})
		}

		if err := addToWield(bdeps, info.Graph, pkggraph.InstallConvenient); err != nil {
			return err
		}

		added[key] = true
		forgeOrder = append(forgeOrder, key)
		toForge[key] = info
		return nil
	}

	//TODO
	//toFetch := make([]pkginfo.PkgInfo)
	toWield := graph.Clone()

	//temporary hack for now
	//TODO maintain full list of packages manually installed in a root
	err = toWield.EnableInstalled()
	if err != nil {
		return fmt.Errorf("Unable to load currently installed packges: %v", err.Error())
	}

	addToWield = func(pkgs []spdl.Dep, g *pkggraph.Graph, itype pkggraph.InstallType) error {
		for _, dep := range pkgs {
			log.Info.Format("Wield %v", dep.String())
			err := g.EnablePackage(dep, itype)
			if err != nil {
				return err
			}
		}

		if err := g.Crunch(); err != nil {
			return err
		}

		for _, node := range g.ToForge() {
			err := addToForge(node.Pkginfo().ToDep())
			if err != nil {
				return err
			}
		}
		return nil
	}

	if isForge {
		happy := true
		for _, pkg := range pkgs {
			if err := addToForge(pkg); err != nil {
				happy = false
				log.Error.Format("Unable to resolve package %v: %v", pkg, err.Error())
			}
		}
		if !happy {
			return fmt.Errorf("Could not resolve package dependencies")
		}
	} else {
		if err := addToWield(pkgs, toWield, itype); err != nil {
			return err
		}
	}

	fmt.Println(color.White.String("Packages to Forge:"))
	for _, info := range toForge {
		fmt.Println(info.Pkginfo.PrettyString())
		fmt.Println(color.White.String("Packages to Wield for %v:", info.Pkginfo.Name))
		for _, pkg := range info.Graph.ToWield() {
			fmt.Println("\t" + pkg.Pkginfo().String())
		}

	}

	fmt.Println(color.White.String("Packages to Wield:"))
	for _, pkg := range toWield.ToWield() {
		fmt.Println(pkg.Pkginfo().String())
	}

	//TODO actually perform stuff

	return nil
}
