package repo

import (
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/helpers/json"
	"github.com/serenitylinux/libspack/pkginfo"
)

type Entry struct {
	Control   control.Control
	Template  string
	Available []pkginfo.PkgInfo
}

// Map<name-version, Tuple<control,pkginfo,hashlist>>
type PkgInstallSetMap map[string]PkgInstallSet

type Repo struct {
	Name        string
	Description string
	//Buildable
	RemoteTemplates string //Templates
	//Installable (pkgset + spakg)
	RemotePackages string //Control + PkgInfo
	Version        string

	//Private NOT SERIALIZED
	entries   map[string][]Entry
	installed *PkgInstallSetMap
}

func Load(filename string) (*Repo, error) {
	var repo Repo
	if err := json.DecodeFile(filename, &repo); err != nil {
		return nil, err
	}
	repo.LoadCaches()
	return &repo, nil
}
