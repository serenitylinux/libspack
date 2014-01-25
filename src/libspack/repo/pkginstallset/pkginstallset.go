package pkginstallset

import (
	"libspack/pkginfo"
	"libspack/control"
	"libspack/hash"
)
import json "libspack/jsonhelper"

type PkgInstallSet struct {
	Control control.Control
	PkgInfo pkginfo.PkgInfo
	Hashes  hash.HashList
}
func (p *PkgInstallSet) ToFile(filename string) error {
	return json.EncodeFile(filename, true, p)
}
func FromFile(filename string) (p *PkgInstallSet, err error) {
	var i PkgInstallSet
	err = json.DecodeFile(filename, &i)
	if err == nil {
		p = &i
	}
	return
}