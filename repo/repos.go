package repo

import (
	"fmt"
	"io/ioutil"

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
		if _, ok := repo.entries[pkgname]; ok {
			return repo, nil
		}
	}
	return nil, fmt.Errorf("Unable to find repo for %v", pkgname)
}

func GetPackageAllVersions(pkgname string) (res []control.Control, repo *Repo) {
	for _, repo = range repos {
		repo.MapByName(pkgname, func(e Entry) {
			res = append(res, e.Control)
		})
		if len(res) != 0 {
			return res, repo
		}
	}
	return nil, nil
}

func GetPackageVersionIteration(pkgname, version string, iteration int) (c *control.Control, repo *Repo) {
	for _, repo = range repos {
		repo.MapByName(pkgname, func(e Entry) {
			if e.Control.Version == version && e.Control.Iteration == iteration {
				c = &e.Control
			}
		})
		if c != nil {
			return c, repo
		}
	}
	return nil, nil
}
func GetPackageVersion(pkgname, version string) (c *control.Control, repo *Repo) {
	for _, repo = range repos {
		repo.MapByName(pkgname, func(e Entry) {
			if e.Control.Version == version {
				if c == nil || e.Control.GreaterThan(*c) {
					c = &e.Control
				}
			}
		})
		if c != nil {
			return c, repo
		}
	}
	return nil, nil
}
func GetPackageLatest(pkgname string) (c *control.Control, repo *Repo) {
	for _, repo = range repos {
		repo.MapByName(pkgname, func(e Entry) {
			if c == nil || e.Control.GreaterThan(*c) {
				c = &e.Control
			}
		})
		if c != nil {
			return c, repo
		}
	}
	return nil, nil
}
func GetPackageInstalledByName(pkgname string, destdir string) (p *PkgInstallSet, repo *Repo) {
	for _, repo = range repos {
		repo.MapInstalledByName(pkgname, destdir, func(installed PkgInstallSet) {
			p = &installed
		})
		if p != nil {
			return p, repo
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
