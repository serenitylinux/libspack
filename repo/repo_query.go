package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/dep"
	"github.com/serenitylinux/libspack/pkginfo"
)

import . "github.com/serenitylinux/libspack/misc"

func (repo *Repo) GetAllControls() ControlMap {
	return *repo.controls
}

func (repo *Repo) GetControls(pkgname string) (control.ControlList, bool) {
	res, e := repo.GetAllControls()[pkgname]
	return res, e
}

func (repo *Repo) GetLatestControl(pkgname string) (*control.Control, bool) {
	c, exists := repo.GetControls(pkgname)
	var res *control.Control = nil

	if exists {
		for _, ctrl := range c {
			if res == nil || res.UUID() < ctrl.UUID() {
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
			if (res == nil || res.UUID() < ctrl.UUID()) && checker(ctrl.Version) {
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
	byUUID := byName[c.UUID()]
	if !exists {
		return "", false
	}
	return byUUID, true
}

func (repo *Repo) GetSpakgOutput(p *pkginfo.PkgInfo) string {
	if !PathExists(SpakgDir + repo.Name) {
		os.MkdirAll(SpakgDir+repo.Name, 0755)
	}
	repo.spakgDir()
	return repo.spakgDir() + fmt.Sprintf("%s.spakg", p.UUID())
}

func (repo *Repo) HasRemoteSpakg(p *pkginfo.PkgInfo) bool {
	_, exists := (*repo.fetchable)[p.UUID()]
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
			if dep.Flags != nil && !dep.Flags.IsSubSet(set.PkgInfo.ParsedFlagStates()) {
				continue
			}
			next = append(next, set)
		}
		candidates = next
	}

	return len(candidates) != 0
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

func (repo *Repo) IsInstalled(p *pkginfo.PkgInfo, destdir string) bool {
	list, err := repo.pkgsInstalledInRoot(destdir)
	if err != nil {
		return false
	}
	_, exists := (*list)[p.UUID()]
	return exists
}
func (repo *Repo) IsAnyInstalled(c *control.Control, destdir string) bool {
	list, err := repo.pkgsInstalledInRoot(destdir)
	if err != nil {
		return false
	}
	for _, pkg := range *list {
		if pkg.Control.UUID() == c.UUID() {
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
		if set.PkgInfo.UUID() == p.UUID() {
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
			if dep == p.Name {
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
				if dep == cur.Name {
					pkgs = append(pkgs, set)
					inner(set.PkgInfo)
				}
			}
		}
	}

	inner(p)

	return pkgs
}
