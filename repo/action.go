package repo

import (
	"fmt"
	"net/url"
	"os"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/hash"
	"github.com/serenitylinux/libspack/helpers/http"
	"github.com/serenitylinux/libspack/pkginfo"
	"github.com/serenitylinux/libspack/spakg"
)
import . "github.com/serenitylinux/libspack/misc"

func (repo *Repo) FetchIfNotCachedSpakg(p pkginfo.PkgInfo) error {
	out := repo.GetSpakgOutput(p)
	if !PathExists(out) {
		src := repo.RemotePackages + "/pkgs/" + url.QueryEscape(fmt.Sprintf("%s.spakg", p))
		log.Info.Format("Fetching %s", src)
		err := http.HttpFetchFileProgress(src, out, true)
		if err != nil {
			os.Remove(out)
		}
		return err
	}
	return nil
}

func (repo *Repo) InstallSpakg(spkg *spakg.Spakg, basedir string) error {
	return repo.Install(spkg.Control, spkg.Pkginfo, spkg.Md5sums, basedir)
}

func (repo *Repo) Install(c control.Control, p pkginfo.PkgInfo, hl hash.HashList, basedir string) error {
	ps := NewPkgIS(&c, &p, hl)
	err := os.MkdirAll(basedir+repo.installedPkgsDir(), 0755)
	if err != nil {
		return err
	}

	err = repo.MapInstalledByName(c.Name, basedir, func(old PkgInstallSet) {
		if old.PkgInfo.String() != p.String() {
			for file, _ := range old.Hashes {
				if _, exists := hl[file]; !exists {
					err := os.RemoveAll(file)
					if err != nil {
						log.Warn.Format("Unable to remove old file %s: %s", file, err)
					}
				}
			}
			repo.MarkRemoved(old.PkgInfo, basedir)
		}
	})

	err = ps.ToFile(repo.installSetFile(p, basedir))
	repo.loadInstalledPackagesList()
	return err
}

func (repo *Repo) MarkRemoved(p *pkginfo.PkgInfo, basedir string) error {
	return os.Remove(repo.installSetFile(*p, basedir))
}

func (repo *Repo) Uninstall(p *pkginfo.PkgInfo, root string) error {
	var err error
	mapErr := repo.MapInstalledByName(p.Name, root, func(inst PkgInstallSet) {
		if inst.PkgInfo.String() != p.String() {
			return
		}

		log.Info.Format("Removing %s", inst.PkgInfo)

		for f, _ := range inst.Hashes {
			log.Debug.Println("Remove: " + root + f)
			err := os.Remove(root + f)
			if err != nil {
				log.Warn.Println(err)
				//Do we return or keep trying?
			}
		}
		err = os.Remove(root + repo.installedPkgsDir() + inst.PkgInfo.String() + ".pkgset")
	})
	if err != nil {
		return err
	}
	return mapErr
}
