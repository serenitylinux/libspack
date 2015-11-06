package repo

import (
	"io/ioutil"
	"net/url"
	"os"
	"regexp"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/helpers/git"
	"github.com/serenitylinux/libspack/helpers/http"
	"github.com/serenitylinux/libspack/helpers/json"
	"github.com/serenitylinux/libspack/pkginfo"
	"github.com/serenitylinux/libspack/spakg"
)

import . "github.com/serenitylinux/libspack/misc"

/*
Repo Dir Management
*/

func (repo *Repo) RefreshRemote() {
	if repo.RemoteTemplates != "" {
		log.Info.Println("Checking remoteTemplates")
		log.Debug.Println(repo.RemoteTemplates)
		cloneRepo(repo.RemoteTemplates, repo.templatesDir(), repo.Name)
	}
	if repo.RemotePackages != "" {
		log.Info.Println("Checking remotePackages")
		log.Debug.Println(repo.RemotePackages)
		cloneRepo(repo.RemotePackages, repo.packagesDir(), repo.Name)
	}

	repo.UpdateCaches()
}

func (repo *Repo) UpdateCaches() {
	repo.entries = make(map[string][]Entry)
	//if we have remote templates
	if repo.RemoteTemplates != "" {
		repo.updateControlsFromTemplates()
		// else if we just have remote controls and prebuilt packages
	} else if repo.RemotePackages != "" {
		repo.updateControlsFromRemote()
	}

	if repo.RemotePackages != "" {
		repo.updatePkgInfosFromRemote()
	}

	repo.loadLocal()

	json.EncodeFile(repo.cacheFile(), repo.entries)
}

func (repo *Repo) LoadCaches() {
	repo.loadCache()
	repo.loadInstalledPackagesList()
}

func cloneRepo(remote string, dir string, name string) {
	switch {
	case GitRegex.MatchString(remote):
		os.MkdirAll(dir, 0755)
		err := git.CloneOrUpdate(remote, dir)
		if err != nil {
			log.Warn.Format("Update repository %s %s failed: %s", name, remote, err)
		}
	case HttpRegex.MatchString(remote):
		os.MkdirAll(dir, 0755)
		listFile := "packages.list"
		err := http.HttpFetchFileProgress(remote+listFile, dir+listFile, false)
		if err != nil {
			log.Warn.Println(err, remote+listFile)
			return
		}

		list := make([]string, 0)
		err = json.DecodeFile(dir+listFile, &list)
		if err != nil {
			log.Warn.Println(err)
			return
		}

		for _, item := range list {
			if !PathExists(dir + item) {
				log.Debug.Format("Fetching %s", item)
				src := remote + "/info/" + url.QueryEscape(item)
				err = http.HttpFetchFileProgress(src, dir+item, false)
				if err != nil {
					log.Warn.Println("Unable to fetch %s: %s", err)
				}
			} else {
				log.Debug.Format("Skipping %s", item)
			}
		}
	case RsyncRegex.MatchString(remote):
		log.Warn.Println("TODO rsync repo")
	default:
		log.Warn.Format("Unknown repository format %s: '%s'", name, remote)
	}
}

func readAll(dir string, regex *regexp.Regexp, todo func(file string)) error {
	if !PathExists(dir) {
		//TODO return errors.New("Unable to access directory")
		return nil
	}

	filelist, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range filelist {
		if regex.MatchString(dir + file.Name()) {
			todo(dir + file.Name())
		}
	}
	return nil
}

func (repo *Repo) addEntry(e Entry) {
	key := e.Control.Name
	if _, ok := repo.entries[key]; !ok {
		log.Debug.Format("Adding: %+v", e)
		repo.entries[key] = []Entry{e}
		return
	}

	foundIndex := -1
	var found Entry
	for i, entry := range repo.entries[key] {
		if entry.Control.String() == e.Control.String() {
			foundIndex = i
			found = entry
			break
		}
	}

	if foundIndex == -1 {
		log.Debug.Format("Adding: %+v", e)
		repo.entries[key] = append(repo.entries[key], e)
		return
	}

	//MERGE

	if e.Template != "" {
		if found.Template != "" {
			log.Warn.Format("Duplicate template for %v: %v vs %v", key, e.Template, found.Template)
		} else {
			found.Template = e.Template
		}
	}

	if len(e.Available) != 0 {
		found.Available = append(found.Available, e.Available...)
	}

	repo.entries[key][foundIndex] = found
}

func (repo *Repo) updateControlsFromTemplates() {
	readFunc := func(file string) {
		c, err := control.FromTemplateFile(file)
		if err != nil {
			log.Warn.Format("Invalid template in repo %s (%s) : %s", repo.Name, file, err.Error())
			return
		}
		repo.addEntry(Entry{Control: c, Template: file})
	}

	err := readAll(repo.templatesDir(), regexp.MustCompile(".*\\.pie$"), readFunc)
	if err != nil {
		log.Warn.Format("Unable to load repo %s's templates: %s", repo.Name, err)
	}
}

//TODO merge with updatePkgiInfosFromRemote
func (repo *Repo) updateControlsFromRemote() {
	readFunc := func(file string) {
		var c control.Control
		err := json.DecodeFile(file, &c)
		if err != nil {
			log.Warn.Format("Invalid control %s in repo %s: %v", file, repo.Name, err.Error())
			return
		}
		repo.addEntry(Entry{Control: c})
	}

	err := readAll(repo.packagesDir(), regexp.MustCompile(".*.control"), readFunc)
	if err != nil {
		log.Warn.Format("Unable to load repo %s's controls: %s", repo.Name, err)
	}
}

func (repo *Repo) updatePkgInfosFromRemote() {
	readFunc := func(file string) {
		var pki pkginfo.PkgInfo
		err := json.DecodeFile(file, &pki)
		if err != nil {
			log.Warn.Format("Invalid pkginfo %s in repo %s: %v", file, repo.Name, err.Error())
			return
		}
		repo.addEntry(Entry{
			//TODO HACK to fake control
			Control: control.Control{
				Name:      pki.Name,
				Version:   pki.Version,
				Iteration: pki.Iteration,
			},
			Available: []pkginfo.PkgInfo{pki},
		})
	}

	err := readAll(repo.packagesDir(), regexp.MustCompile(".*.pkginfo"), readFunc)
	if err != nil {
		log.Warn.Format("Unable to load repo %s's controls: %s", repo.Name, err)
	}
}

//TODO rewrite
func (repo *Repo) loadLocal() {
	readFunc := func(file string) {
		tmpDir, _ := ioutil.TempDir(os.TempDir(), "repo")
		defer os.RemoveAll(tmpDir)

		pkg, err := spakg.FromFile(file, &tmpDir)
		if err != nil {
			log.Warn.Format("Error loading %v: %v", file, err.Error())
			return
		}
		if file != repo.GetSpakgOutput(pkg.Pkginfo) {
			log.Warn.Format("Error loading %v: %v", file, "Mismatched checksums: "+pkg.Pkginfo.String())
			return
		}

		repo.addEntry(Entry{
			Control:   pkg.Control,
			Available: []pkginfo.PkgInfo{pkg.Pkginfo},
		})
	}

	err := readAll(repo.spakgDir(), regexp.MustCompile(".*.spakg"), readFunc)
	if err != nil {
		log.Warn.Format("Unable to load repo %s's controls: %s", repo.Name, err)
	}
	return
}

func (repo *Repo) loadCache() {
	log.Debug.Format("Loading cache for %s", repo.Name)
	cf := repo.cacheFile()
	if PathExists(cf) {
		err := json.DecodeFile(cf, &repo.entries)
		if err != nil {
			log.Warn.Format("Could not load cache for repo %s: %s", repo.Name, err)
		}
	}
}

func (repo *Repo) loadInstalledPackagesList() {
	log.Debug.Format("Loading installed packages for %s", repo.Name)

	dir := repo.installedPkgsDir()

	if !PathExists(dir) {
		os.MkdirAll(dir, 0755)
		return
	}

	list, err := installedPackageList(dir)
	if err != nil {
		log.Error.Format("Unable to load repo %s's installed packages: %s", repo.Name, err)
		log.Warn.Println("This is a REALLY bad thing!")
	}
	repo.installed = list
}

func installedPackageList(dir string) (*PkgInstallSetMap, error) {
	list := make(PkgInstallSetMap)

	readFunc := func(file string) {
		ps, err := PkgISFromFile(file)

		if err != nil {
			log.Error.Format("Invalid pkgset %s: %s", file, err)
			log.Warn.Println("This is a REALLY bad thing!")
			return
		}

		list[ps.PkgInfo.String()] = *ps
	}

	err := readAll(dir, regexp.MustCompile(".*.pkgset"), readFunc)
	return &list, err
}
