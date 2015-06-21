package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/pkginfo"
)

import . "github.com/serenitylinux/libspack/misc"

func (repo *Repo) MapTemplatesByName(name string, fn func(string, control.Control)) {
	for _, cs := range *repo.controls {
		if len(cs) == 0 {
			continue
		}
		if cs[0].Name != name {
			continue
		}
		for _, c := range cs {
			if ts, ok := (*repo.templateFiles)[c.Name]; ok {
				if t, ok := ts[c.String()]; ok {
					fn(t, c)
				}
			}
		}
	}

}

func (repo *Repo) MapAvailableByName(name string, fn func(control.Control, pkginfo.PkgInfo)) {
	for _, ps := range *repo.fetchable {
		if len(ps) == 0 {
			continue
		}
		if ps[0].Name != name {
			continue
		}
		for _, p := range ps {
			if cs, ok := (*repo.controls)[p.Name]; ok {
				for _, c := range cs {
					if c.Version == p.Version {
						fn(c, p)
						break
					}
				}
			} else {
				panic("info without template")
			}
		}
	}
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

//TODO === refactor below this line ====

func (repo *Repo) GetAllNames() []string {
	res := make([]string, 0, len(*repo.controls))
	for name := range *repo.controls {
		res = append(res, name)
	}
	return res
}

func (repo *Repo) GetAllControls() ControlMap {
	return *repo.controls
}

func (repo *Repo) GetControls(pkgname string) ([]control.Control, bool) {
	res, e := repo.GetAllControls()[pkgname]
	return res, e
}

func (repo *Repo) GetLatestControl(pkgname string) (*control.Control, bool) {
	c, exists := repo.GetControls(pkgname)
	var res *control.Control = nil

	if exists {
		for _, ctrl := range c {
			if res == nil || res.String() < ctrl.String() {
				res = &ctrl
			}
		}
	}
	return res, res != nil
}

func (repo *Repo) GetPackageByVersionChecker(pkgname string, checker func(string) bool) *control.Control {
	c, exists := repo.GetControls(pkgname)
	var res *control.Control = nil

	if exists {
		for _, ctrl := range c {
			if (res == nil || res.String() < ctrl.String()) && checker(ctrl.Version) {
				res = &ctrl
			}
		}
	}
	return res
}

func (repo *Repo) GetAllTemplates() TemplateFileMap {
	return *repo.templateFiles
}

func (repo *Repo) GetTemplateByControl(c *control.Control) (string, bool) {
	byName, exists := repo.GetAllTemplates()[c.Name]
	if !exists {
		return "", false
	}
	byString := byName[c.String()]
	if !exists {
		return "", false
	}
	return byString, true
}

func (repo *Repo) GetSpakgOutput(p *pkginfo.PkgInfo) string {
	if !PathExists(SpakgDir + repo.Name) {
		os.MkdirAll(SpakgDir+repo.Name, 0755)
	}
	repo.spakgDir()
	return repo.spakgDir() + fmt.Sprintf("%s.spakg", p)
}

func (repo *Repo) HasRemoteSpakg(p *pkginfo.PkgInfo) bool {
	_, exists := (*repo.fetchable)[p.String()]
	return exists
}
func (repo *Repo) HasLocalSpakg(p *pkginfo.PkgInfo) bool {
	return PathExists(repo.GetSpakgOutput(p))
}

func (repo *Repo) HasSpakg(p *pkginfo.PkgInfo) bool {
	return repo.HasLocalSpakg(p) || repo.HasRemoteSpakg(p)
}

func (repo *Repo) HasAnySpakg(c *control.Control) bool {
	for _, plist := range *repo.fetchable {
		for _, p := range plist {
			if p.InstanceOf(c) {
				return true
			}
		}
	}

	return false
}

func (repo *Repo) HasTemplate(c *control.Control) bool {
	_, exists := repo.GetTemplateByControl(c)
	return exists
}

/*  TODO
func (repo *Repo) AnyInstalled(pkg string, deps dep.DepList, destdir string) bool {
	candidates := make([]PkgInstallSet, 0)
	list, err := repo.pkgsInstalledInRoot(destdir)
	if err != nil {
		return false
	}

	for _, set := range *list {
		if set.Control.Name == pkg {
			candidates = append(candidates, set)
		}
	}

	for _, dep := range deps {
		next := make([]PkgInstallSet, 0)
		for _, set := range candidates {
			if dep.Version1 != nil && !dep.Version1.Accepts(set.PkgInfo.Version) {
				continue
			}
			if dep.Version2 != nil && !dep.Version2.Accepts(set.PkgInfo.Version) {
				continue
			}
			if dep.Flags != nil && !dep.Flags.IsSubsetOf(set.PkgInfo.FlagStates) {
				continue
			}
			next = append(next, set)
		}
		candidates = next
	}

	return len(candidates) != 0
}*/

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

func (repo *Repo) IsInstalled(p *pkginfo.PkgInfo, destdir string) bool {
	list, err := repo.pkgsInstalledInRoot(destdir)
	if err != nil {
		return false
	}
	_, exists := (*list)[p.String()]
	return exists
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

func (repo *Repo) GetAllInstalled() []PkgInstallSet {
	res := make([]PkgInstallSet, 0)
	for _, i := range *repo.installed {
		res = append(res, i)
	}
	return res
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

func (repo *Repo) GetInstalled(p *pkginfo.PkgInfo, destdir string) *PkgInstallSet {
	list, err := repo.pkgsInstalledInRoot(destdir)
	if err != nil {
		return nil
	}
	for _, set := range *list {
		if set.PkgInfo.String() == p.String() {
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
