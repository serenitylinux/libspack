package repo

import (
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/helpers/json"
	"github.com/serenitylinux/libspack/pkginfo"
)

//Sorted by pkgversion
type ControlMap map[string][]control.Control

// Map<name, map<version>>
type TemplateFileMap map[string]map[string]string

// Map<name-version, List<PkgInfo>>
type PkgInfoMap map[string][]pkginfo.PkgInfo

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
	controls      *ControlMap
	templateFiles *TemplateFileMap
	fetchable     *PkgInfoMap
	installed     *PkgInstallSetMap
}

func Load(filename string) (*Repo, error) {
	var repo Repo
	if err := json.DecodeFile(filename, &repo); err != nil {
		return nil, err
	}
	repo.LoadCaches()
	return &repo, nil
}
