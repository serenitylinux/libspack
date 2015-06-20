package repo

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/cam72cam/go-lumberjack/color"
	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/misc"
	"github.com/serenitylinux/libspack/pkginfo"
)

const reposDir = "/etc/spack/repos/"

type RepoList map[string]*Repo

var repos RepoList

func init() {
	LoadRepos()
}

func LoadRepos() error {
	repos = make(RepoList)
	files, err := ioutil.ReadDir(reposDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		fAbs := reposDir + f.Name()
		r, err := Load(fAbs)
		if err != nil {
			return err
		}

		repos[r.Name] = r
	}
	return nil
}

func RefreshRepos(notRemote bool) {
	log.Info.Println()
	for _, repo := range repos {
		log.Info.Println("Refreshing ", repo.Name)
		misc.LogBar(log.Info, color.Brown)
		if notRemote {
			repo.UpdateCaches()
		} else {
			repo.RefreshRemote()
		}
		misc.PrintSuccess()
	}
}

func GetAllRepos() RepoList {
	return repos
}

func GetRepoFor(pkgname string) (*Repo, error) {
	for _, repo := range repos {
		for _, name := range repo.GetAllNames() {
			if name == pkgname {
				return repo, nil
			}
		}
	}
	return nil, fmt.Errorf("Unable to find repo for %v", pkgname)
}

func GetPackageAllVersions(pkgname string) ([]control.Control, *Repo) {
	for _, repo := range repos {
		cl, exists := repo.GetControls(pkgname)
		if exists {
			return cl, repo
		}
	}
	return nil, nil
}

func GetPackageVersionIteration(pkgname, version, iteration string) (*control.Control, *Repo) {
	pkgs, repo := GetPackageAllVersions(pkgname)
	itri, e := strconv.Atoi(iteration)
	if e != nil {
		log.Warn.Println(e)
		return nil, nil
	}
	var ctrl *control.Control
	for _, ver := range pkgs {
		if ver.Version == version {
			if itri == ver.Iteration {
				ctrl = &ver
				break
			}
		}
	}
	if ctrl == nil {
		return nil, nil
	} else {
		return ctrl, repo
	}
}
func GetPackageVersion(pkgname, version string) (*control.Control, *Repo) {
	pkgs, repo := GetPackageAllVersions(pkgname)
	var ctrl *control.Control
	for _, ver := range pkgs {
		if ver.Version == version {
			if ctrl == nil || ctrl.Iteration < ver.Iteration {
				ctrl = &ver
			}
		}
	}
	if ctrl == nil {
		return nil, nil
	} else {
		return ctrl, repo
	}
}
func GetPackageLatest(pkgname string) (*control.Control, *Repo) {
	for _, repo := range repos {
		c, exists := repo.GetLatestControl(pkgname)
		if exists {
			return c, repo
		}
	}
	return nil, nil
}
func GetPackageInstalledByName(pkgname string, destdir string) (*PkgInstallSet, *Repo) {
	for _, repo := range repos {
		c := repo.GetInstalledByName(pkgname, destdir)
		if c != nil {
			return c, repo
		}
	}
	return nil, nil
}
func UninstallList(p *pkginfo.PkgInfo) []PkgInstallSet {
	res := make([]PkgInstallSet, 0)
	for _, repo := range repos {
		res = append(res, repo.UninstallList(p)...)
	}
	return res
}
func RdepList(p *pkginfo.PkgInfo) []PkgInstallSet {
	res := make([]PkgInstallSet, 0)
	for _, repo := range repos {
		res = append(res, repo.RdepList(p)...)
	}
	return res
}
