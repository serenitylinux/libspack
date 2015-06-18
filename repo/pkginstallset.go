package repo

import (
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/hash"
	"github.com/serenitylinux/libspack/helpers/json"
	"github.com/serenitylinux/libspack/pkginfo"
)

type PkgInstallSet struct {
	Control *control.Control
	PkgInfo *pkginfo.PkgInfo
	Hashes  hash.HashList
}

func NewPkgIS(c *control.Control, p *pkginfo.PkgInfo, hash hash.HashList) *PkgInstallSet {
	return &PkgInstallSet{c, p, hash}
}
func (p *PkgInstallSet) ToFile(filename string) error {
	return json.EncodeFile(filename, p)
}
func PkgISFromFile(filename string) (*PkgInstallSet, error) {
	var i PkgInstallSet
	if err := json.DecodeFile(filename, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

func (repo *Repo) installSetFile(p pkginfo.PkgInfo, basedir string) string {
	return basedir + repo.installedPkgsDir() + p.String() + ".pkgset"
}
