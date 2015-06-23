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
}

func (repo *Repo) LoadCaches() {
	repo.loadControlCache()
	repo.loadPkgInfoCache()
	repo.loadTemplateListCache()
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
			todo(dir + "/" + file.Name())
		}
	}
	return nil
}

//Will also populate template list
func (repo *Repo) updateControlsFromTemplates() {
	//Generates new list and writes to cache
	list := make(ControlMap)

	dir := repo.templatesDir()

	readFunc := func(file string) {
		c, err := control.FromTemplateFile(file)

		if err != nil {
			log.Warn.Format("Invalid template in repo %s (%s) : %s", repo.Name, file, err.Error())
			return
		}

		// Initialize list of controls for current name if nessesary
		if _, exists := list[c.Name]; !exists {
			list[c.Name] = make([]control.Control, 0)
		}

		if _, exists := (*repo.templateFiles)[c.Name]; !exists {
			(*repo.templateFiles)[c.Name] = make(map[string]string)
		}

		(*repo.templateFiles)[c.Name][c.String()] = file
		list[c.Name] = append(list[c.Name], c)
	}

	err := readAll(dir, regexp.MustCompile(".*\\.pie"), readFunc)

	if err != nil {
		log.Warn.Format("Unable to load repo %s's templates: %s", repo.Name, err)
		return
	}

	repo.controls = &list
	json.EncodeFile(repo.controlCacheFile(), repo.controls)
	json.EncodeFile(repo.templateListCacheFile(), repo.templateFiles)
}

func (repo *Repo) updateControlsFromRemote() {
	// finds all files in remote dir and writes to cache
	list := make(ControlMap)

	readFunc := func(file string) {
		var c control.Control
		err := json.DecodeFile(file, &c)
		if err != nil {
			log.Warn.Format("Invalid control %s in repo %s: %v", file, repo.Name, err.Error())
			return
		}

		if _, exists := list[c.Name]; !exists {
			list[c.Name] = make([]control.Control, 0)
		}
		list[c.Name] = append(list[c.Name], c)
	}

	err := readAll(repo.packagesDir(), regexp.MustCompile(".*.control"), readFunc)

	if err != nil {
		log.Warn.Format("Unable to load repo %s's controls: %s", repo.Name, err)
		return
	}

	repo.controls = &list
	json.EncodeFile(repo.controlCacheFile(), repo.controls)
}

//TODO rewrite
func (repo *Repo) LoadLocal() error {
	if repo.local != nil && len(*repo.local) != 0 {
		return nil
	}
	dir := "/var/cache/spack/spakg/" + repo.Name

	println("loading caches " + dir)

	list := make(PkgInfoMap)

	if !PathExists(dir) {
		println("whaaa")
		//TODO return errors.New("Unable to access directory")
		return nil
	}

	filelist, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range filelist {
		tmpDir, _ := ioutil.TempDir(os.TempDir(), "wield")
		defer os.RemoveAll(tmpDir)

		pkg, err := spakg.FromFile(dir+"/"+file.Name(), &tmpDir)
		if err != nil {
			log.Warn.Format("Error loading %v: %v", dir+"/"+file.Name(), err.Error())
			continue
		}
		if file.Name() != pkg.Pkginfo.String()+".spakg" {
			log.Warn.Format("Error loading %v: %v", dir+"/"+file.Name(), "Mismatched checksums: "+pkg.Pkginfo.String())
			continue
		}

		pki := pkg.Pkginfo
		key := pki.String()
		if _, exists := list[key]; !exists {
			list[key] = make([]pkginfo.PkgInfo, 0)
		}
		list[key] = append(list[key], pki)

		c := pkg.Control

		if _, exists := (*repo.controls)[c.Name]; !exists {
			(*repo.controls)[c.Name] = make([]control.Control, 0)
		}
		(*repo.controls)[c.Name] = append((*repo.controls)[c.Name], c)
	}

	println("found ", len(list))

	repo.local = &list
	return nil
}

func (repo *Repo) updatePkgInfosFromRemote() {
	//Generates new list and writes to cache
	list := make(PkgInfoMap)

	readFunc := func(file string) {
		var pki pkginfo.PkgInfo
		err := json.DecodeFile(file, &pki)

		if err != nil {
			log.Warn.Format("Invalid pkginfo %s in repo %s: %v", file, repo.Name, err.Error())
			return
		}

		key := pki.String()
		if _, exists := list[key]; !exists {
			list[key] = make([]pkginfo.PkgInfo, 0)
		}
		list[key] = append(list[key], pki)
	}

	err := readAll(repo.packagesDir(), regexp.MustCompile(".*.pkginfo"), readFunc)
	if err != nil {
		log.Warn.Format("Unable to load repo %s's controls: %s", repo.Name, err)
		return
	}

	repo.fetchable = &list
	json.EncodeFile(repo.pkgInfoCacheFile(), repo.fetchable)
}

func (repo *Repo) loadControlCache() {
	log.Debug.Format("Loading controls for %s", repo.Name)
	list := make(ControlMap)
	cf := repo.controlCacheFile()
	if PathExists(cf) {
		err := json.DecodeFile(cf, &list)
		if err != nil {
			log.Warn.Format("Could not load control cache for repo %s: %s", repo.Name, err)
		}
	}
	repo.controls = &list
}

func (repo *Repo) loadPkgInfoCache() {
	log.Debug.Format("Loading pkginfos for %s", repo.Name)
	list := make(PkgInfoMap)
	pif := repo.pkgInfoCacheFile()
	if PathExists(pif) {
		err := json.DecodeFile(pif, &list)
		if err != nil {
			log.Warn.Format("Could not load pkginfo cache for repo %s: %s", repo.Name, err)
		}
	}
	repo.fetchable = &list
}

func (repo *Repo) loadTemplateListCache() {
	log.Debug.Format("Loading templates for %s", repo.Name)
	list := make(TemplateFileMap)
	tlf := repo.templateListCacheFile()
	if PathExists(tlf) {
		err := json.DecodeFile(tlf, &list)
		if err != nil {
			log.Warn.Format("Could not load template list cache for repo %s: %s", repo.Name, err)
		}
	}
	repo.templateFiles = &list
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