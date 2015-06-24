package libspack

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/cam72cam/go-lumberjack/color"
	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/forge"
	"github.com/serenitylinux/libspack/misc"
	"github.com/serenitylinux/libspack/crunch"
	"github.com/serenitylinux/libspack/pkginfo"
	"github.com/serenitylinux/libspack/repo"
	"github.com/serenitylinux/libspack/spakg"
	"github.com/serenitylinux/libspack/spdl"
	"github.com/serenitylinux/libspack/wield"
)

//TODO: reinstall

func Forge(pkgs []spdl.Dep, root string, ignoreBDeps bool) error {
	return buildGraphs(pkgs, true, root, ignoreBDeps, false, crunch.InstallConvenient)
}

func Wield(pkgs []spdl.Dep, root string, reinstall bool, itype crunch.InstallType) error {
	return buildGraphs(pkgs, false, root, false, reinstall, itype)
}

func buildGraphs(pkgs []spdl.Dep, isForge bool, root string, ignoreBDeps bool, reinstall bool, itype crunch.InstallType) error {
	type forgeInfo struct {
		Graph *crunch.Graph
		Root  string

		Pkginfo  *pkginfo.PkgInfo
		Control  *control.Control
		Repo     *repo.Repo
		Template string
	}

	graph, err := crunch.NewGraph(root, repo.GetAllRepos())
	if err != nil {
		return fmt.Errorf("Unable to load package graph: %v", err.Error())
	}

	var addToForge func(spdl.Dep) error
	var addToWield func([]spdl.Dep, *crunch.Graph, crunch.InstallType) error

	toForge := make(map[string]*forgeInfo, 0)
	added := make(map[string]bool)
	forgeOrder := make([]string, 0)
	toRemove := make([]string, 0)

	addToForge = func(pkg spdl.Dep) (err error) {
		log.Info.Format("Forge %v", pkg.String())
		info := &forgeInfo{
			Graph: graph.Clone(),
		}
		info.Root, err = ioutil.TempDir(root+os.TempDir(), "forge")
		if err != nil {
			return err
		}
		info.Root += "/"
		toRemove = append(toRemove, info.Root)
		log.Debug.Format("Root: %v", info.Root)
		info.Graph.ChangeRoot(info.Root)

		info.Repo, err = repo.GetRepoFor(pkg.Name)
		if err != nil {
			return err
		}
		info.Repo.MapTemplatesByName(pkg.Name, func(t string, c control.Control) {
			if info.Control == nil || spdl.NewVersion(spdl.GT, info.Control.Version).Accepts(c.Version) {
				info.Control = &c
				info.Template = t
			}
		})

		if info.Control == nil {
			return fmt.Errorf("Unable to find package %v", pkg.Name)
		}

		var flags spdl.FlatFlagList
		if pkg.Flags != nil {
			flags, err = pkg.Flags.WithDefaults(spdl.NewFlatFlagList(0))
			if err != nil {
				return err
			}
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

		libc, _ := spdl.ParseDep("libc(+dev)")
		if err := addToWield([]spdl.Dep{{Name: "base"}, libc}, info.Graph, crunch.InstallLatestBin); err != nil {
			return err
		}

		if !ignoreBDeps {
			bdeps := make([]spdl.Dep, 0)
			for _, bdep := range info.Control.Bdeps {
				if bdep.Condition != nil && !bdep.Condition.Enabled(info.Pkginfo.FlagStates) {
					continue //Not enabled
				}
				defaults := spdl.NewFlatFlagList(0)
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

			if err := addToWield(bdeps, info.Graph, crunch.InstallLatestBin); err != nil {
				return err
			}
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

	addToWield = func(pkgs []spdl.Dep, g *crunch.Graph, itype crunch.InstallType) error {
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

	if len(toForge) != 0 {
		fmt.Println(color.White.String("Packages to Forge:"))
		for _, info := range toForge {
			fmt.Println(info.Pkginfo.PrettyString())
			if len(info.Graph.ToWield()) != 0 {
				fmt.Println(color.White.String("Packages to Wield for ", info.Pkginfo.PrettyString()))
				for _, pkg := range sortp(info.Graph.ToWield()) {
					fmt.Println("\t" + pkg.Pkginfo().PrettyString())
				}
			}
		}
	}

	if len(toWield.ToWield()) != 0 {
		fmt.Println(color.White.String("Packages to Wield:"))
		for _, pkg := range toWield.ToWield() {
			fmt.Println(pkg.Pkginfo().PrettyString())
		}
	}

	if len(toWield.ToWield()) == 0 && len(toForge) == 0 {
		log.Info.Println("Nothing to do")
		return nil
	}

	if !misc.AskYesNo("Do you wish to continue?", true) {
		return nil
	}

	//TODO interactive mode
	isInteractive := false
	if len(toForge) != 0 {
		for _, info := range toForge {
			log.Info.Format("Installing bdeps for %s", info.Pkginfo.PrettyString())

			if len(info.Graph.ToWield()) != 0 {
				if err := wieldGraph(info.Graph.ToWield(), info.Root); err != nil {
					return err
				}
			}

			log.Info.Format("Forging %s", info.Pkginfo.PrettyString())
			spakgFile := info.Repo.GetSpakgOutput(*info.Pkginfo)
			err = forge.Forge(info.Template, spakgFile, info.Root, info.Pkginfo.FlagStates, false, isInteractive)
			if err != nil {
				return err
			}
		}
	}

	if len(toWield.ToWield()) != 0 {
		err := wieldGraph(toWield.ToWield(), root)
		if err != nil {
			return err
		}
	}

	misc.PrintSuccess()

	return nil
}

func wieldGraph(nodes []*crunch.Node, root string) error {
	type pkgset struct {
		spkg *spakg.Spakg
		repo *repo.Repo
		file string
	}
	spkgs := make([]pkgset, 0)

	//Fetch Packages
	for _, pkg := range nodes {
		pkginfo := pkg.Pkginfo()
		err := pkg.Repo.FetchIfNotCachedSpakg(pkginfo)
		if err != nil {
			return err
		}

		pkgfile := pkg.Repo.GetSpakgOutput(pkginfo)
		spkg, err := spakg.FromFile(pkgfile, nil)
		if err != nil {
			return err
		}

		spkgs = append(spkgs, pkgset{spkg, pkg.Repo, pkgfile})
	}
	log.Info.Println()

	//Preinstall
	for _, pkg := range spkgs {
		wield.PreInstall(pkg.spkg, root)
	}
	log.Debug.Println()

	//Install
	for _, pkg := range spkgs {
		err := wield.ExtractCheckCopy(pkg.file, root)

		if err != nil {
			return err
		}

		pkg.repo.InstallSpakg(pkg.spkg, root)
	}
	log.Debug.Println()
	if len(spkgs) != 0 {
		wield.Ldconfig(root)
	}

	//PostInstall
	for _, pkg := range spkgs {
		wield.PostInstall(pkg.spkg, root)
	}
	log.Info.Println()

	return nil
}

func sortp(orig []*crunch.Node) (nl []*crunch.Node) {
	strs := make([]string, 0, len(orig))
	for _, pkg := range orig {
		strs = append(strs, pkg.Name)
	}
	sort.Strings(strs)
	for _, str := range strs {
		for _, pkg := range orig {
			if pkg.Name == str {
				nl = append(nl, pkg)
				break
			}
		}
	}
	return nl
}
