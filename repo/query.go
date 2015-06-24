package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/pkginfo"
)

import . "github.com/serenitylinux/libspack/misc"

func (repo *Repo) Map(fn func(Entry)) {
	for _, set := range repo.entries {
		for _, entry := range set {
			fn(entry)
		}
	}
}

func (repo *Repo) MapWithName(fn func(name string, entries []Entry)) {
	for name, entries := range repo.entries {
		fn(name, entries)
	}
}

func (repo *Repo) MapByName(name string, fn func(Entry)) {
	if entries, ok := repo.entries[name]; ok {
		for _, entry := range entries {
			fn(entry)
		}
	}
}

func (repo *Repo) MapTemplatesByName(name string, fn func(string, control.Control)) {
	if entries, ok := repo.entries[name]; ok {
		for _, entry := range entries {
			if entry.Template != "" {
				fn(entry.Template, entry.Control)
			}
		}
	}
}

func (repo *Repo) MapAvailableByName(name string, fn func(control.Control, pkginfo.PkgInfo)) {
	if entries, ok := repo.entries[name]; ok {
		for _, entry := range entries {
			for _, pki := range entry.Available {
				fn(entry.Control, pki)
			}
		}
	}
}

func (repo *Repo) MapInstalled(root string, fn func(p PkgInstallSet)) error {
	list, err := repo.pkgsInstalledInRoot(root)
	if err != nil {
		return err
	}
	for _, pkg := range *list {
		fn(pkg)
	}
	return nil
}

func (repo *Repo) MapInstalledByName(root, name string, fn func(p PkgInstallSet)) error {
	list, err := repo.pkgsInstalledInRoot(root)
	if err != nil {
		return err
	}
	for _, pkg := range *list {
		if pkg.Control.Name == name {
			fn(pkg)
		}
	}
	return nil
}

func (repo *Repo) GetSpakgOutput(p pkginfo.PkgInfo) string {
	if !PathExists(SpakgDir + repo.Name) {
		os.MkdirAll(SpakgDir+repo.Name, 0755)
	}
	repo.spakgDir()
	return repo.spakgDir() + fmt.Sprintf("%s.spakg", p)
}

func (repo *Repo) IsAnyInstalled(c *control.Control, destdir string) bool {
	list, err := repo.pkgsInstalledInRoot(destdir)
	if err != nil {
		return false
	}
	for _, pkg := range *list {
		if pkg.Control.String() == c.String() {
			return true
		}
	}
	return false
}
func (repo *Repo) GetInstalledByName(name string, destdir string) *PkgInstallSet {
	list, err := repo.pkgsInstalledInRoot(destdir)
	if err != nil {
		return nil
	}
	for _, set := range *list {
		if set.PkgInfo.Name == name {
			return &set
		}
	}
	return nil
}

// TODO actually check if that dep is enabled or not in the pkginfo
func (repo *Repo) RdepList(p *pkginfo.PkgInfo) []PkgInstallSet {
	pkgs := make([]PkgInstallSet, 0)

	for _, set := range *repo.installed {
		for _, dep := range set.Control.Deps {
			if dep.Name == p.Name {
				pkgs = append(pkgs, set)
			}
		}
	}

	return pkgs
}

// TODO actually check if that dep is enabled or not in the pkginfo
func (repo *Repo) UninstallList(p *pkginfo.PkgInfo) []PkgInstallSet {
	pkgs := make([]PkgInstallSet, 0)

	var inner func(*pkginfo.PkgInfo)

	inner = func(cur *pkginfo.PkgInfo) {
		for _, pkg := range pkgs {
			if pkg.Control.Name == cur.Name {
				return
			}
		}

		for _, set := range *repo.installed {
			for _, dep := range set.Control.Deps {
				if dep.Name == cur.Name {
					pkgs = append(pkgs, set)
					inner(set.PkgInfo)
				}
			}
		}
	}

	inner(p)

	return pkgs
}

var cachedInstalledRoots = make(map[string]*PkgInstallSetMap)

func (repo *Repo) pkgsInstalledInRoot(destdir string) (*PkgInstallSetMap, error) {
	if filepath.Clean(destdir) == "/" {
		return repo.installed, nil
	} else {
		if list, ok := cachedInstalledRoots[destdir]; ok {
			return list, nil
		}
		list, err := installedPackageList(destdir + InstallDir + repo.Name + "/")
		if err != nil {
			return nil, err
		}
		cachedInstalledRoots[destdir] = list
		return list, nil
	}
}
