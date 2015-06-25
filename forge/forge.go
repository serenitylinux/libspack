package forge

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/helpers/git"
	"github.com/serenitylinux/libspack/helpers/http"
	"github.com/serenitylinux/libspack/pkginfo"
	"github.com/serenitylinux/libspack/spakg"
	"github.com/serenitylinux/libspack/spdl"
)

import . "github.com/serenitylinux/libspack/misc"
import . "github.com/serenitylinux/libspack/hash"

const (
	dest = "dest"
	src  = "src"
)

type forgeInfo struct {
	template string
	outfile  string
	root     string
	workdir  string

	control     control.Control
	states      spdl.FlatFlagList
	test        bool
	interactive bool
}

func Forge(template, outfile, root string, states spdl.FlatFlagList, test bool, interactive bool) error {
	c, err := control.FromTemplateFile(template)
	if err != nil {
		return err
	}

	info := forgeInfo{
		template:    template,
		outfile:     outfile,
		root:        root,
		workdir:     "/forge/",
		control:     c,
		states:      states,
		test:        test,
		interactive: interactive,
	}

	return forge(info)
}
func forge(info forgeInfo) error {
	err := os.MkdirAll(info.root+info.workdir, 0755)
	if err != nil {
		return err
	}
	//defer os.RemoveAll(info.root + info.workdir)

	def := filepath.Dir(info.template) + "/default"
	if _, err := os.Stat(def); err == nil {
		log.Info.Format("Including repo default: %v", def)
		cmd := exec.Command("cp", def, info.root+info.workdir)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("cp", info.template, info.root+info.workdir)
	err = cmd.Run()
	if err != nil {
		return err
	}

	os.Mkdir(info.root+info.workdir+dest, 0755)
	os.Mkdir(info.root+info.workdir+src, 0755)

	OnError := func(err error) error {
		if info.interactive {
			log.Error.Println(err)
			log.Info.Println("Dropping you to a shell")
			InDir(info.root+info.workdir, func() {
				cmd := exec.Command("bash")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				cmd.Run()
			})
		}
		return err
	}

	err = FetchPkgSrc(info)
	if err != nil {
		return OnError(err)
	}

	err = runParts(info)
	if err != nil {
		return OnError(err)
	}

	err = StripPackage(info)
	if err != nil {
		return OnError(err)
	}

	err = BuildPackage(info)
	if err != nil {
		return OnError(err)
	}

	return nil
}

func extractPkgSrc(srcPath string, outDir string) error {
	tarRegex := regexp.MustCompile(".*\\.(tar|tgz).*")
	zipRegex := regexp.MustCompile(".*\\.zip")
	var cmd *exec.Cmd
	switch {
	case tarRegex.MatchString(srcPath):
		cmd = exec.Command("tar", "-xvf", srcPath, "-C", outDir)
	case zipRegex.MatchString(srcPath):
		cmd = exec.Command("unzip", srcPath, "-d", outDir)
	default:
		return errors.New("Unknown archive type: " + outDir)
	}
	return RunCommand(cmd, log.Debug, os.Stderr)
}

func FetchPkgSrc(info forgeInfo) error {
	Header("Fetching Source")

	for _, url := range info.control.Src {
		if url == "" { //Hack while we are still using dumb bash str lists for urls
			continue
		}
		gitRegex := regexp.MustCompile("(.*\\.git|git://)")
		httpRegex := regexp.MustCompile("(http|https)://.*")
		ftpRegex := regexp.MustCompile("ftp://.*")

		name := path.Base(url)
		file := info.root + info.workdir + "/" + name
		srcdir := info.root + info.workdir + src
		switch {
		case gitRegex.MatchString(url):
			log.Debug.Format("Fetching '%s' with git", url)
			dir := srcdir + name

			dir = strings.Replace(dir, ".git", "", 1)
			err := os.Mkdir(dir, 0755)
			if err != nil {
				return err
			}

			err = git.Clone(url, dir)
			if err != nil {
				return err
			}

		case ftpRegex.MatchString(url):
			log.Debug.Format("Fetching '%s'", url)
			err := RunCommandToStdOutErr(exec.Command("wget", url, "-O", file))
			if err != nil {
				return err
			}

			err = extractPkgSrc(file, srcdir)
			if err != nil {
				return err
			}

		case httpRegex.MatchString(url):
			log.Debug.Format("Fetching '%s' with http", url)

			err := http.HttpFetchFileProgress(url, file, log.Debug.IsEnabled())
			if err != nil {
				return err
			}

			err = extractPkgSrc(file, srcdir)
			if err != nil {
				return err
			}

		default:
			return errors.New(fmt.Sprintf("Unknown url format '%s', cannot continue", url))
		}
	}
	PrintSuccess()
	return nil
}

func envString(env map[string]string) string {
	var parts []string
	for k, v := range env {
		parts = append(parts, fmt.Sprintf("export %v=%v", k, v))
	}
	return strings.Join(parts, "\n")
}

func runPart(part, action string, info forgeInfo, env map[string]string) error {
	var flagstuff string
	for _, fl := range info.states.Slice() {
		flagstuff += fmt.Sprintf("flag_%s=%t \n", fl.Name, fl.Enabled)
	}

	template := info.workdir + path.Base(info.template)
	defaults := info.workdir + "default"

	forge_helper := `
		` + envString(env) + `

		if ! [ -d /dev/ ]; then
			mkdir /dev;
		fi

		mkdir -p /etc/
		echo "nameserver 8.8.8.8" > /etc/resolv.conf

		function none {
			return 0
		}
		
		function default {
			` + action + `
		}
		
		` + flagstuff + `
		
		source ` + template + `
		
		if [ -f ` + defaults + ` ]; then
			source ` + defaults + `
		fi

		cd ` + info.workdir + src + `/$srcdir
		
		set +e 
		declare -f ` + part + ` > /dev/null
		exists=$?
		set -e
		
		if [ $exists -ne 0 ]; then
			default
		else
			` + part + `
		fi`

	Header("Running " + part)

	bash := exec.Command("bash", "-ce", forge_helper)
	if filepath.Clean(info.root) != "/" {
		if _, err := exec.LookPath("chroot"); err == nil {
			bash.Args = append([]string{info.root}, bash.Args...)
			bash = exec.Command("chroot", bash.Args...)
		} else if _, err := exec.LookPath("systemd-nspawn"); err == nil {
			bash.Args = append([]string{"-D", info.root}, bash.Args...)
			bash = exec.Command("systemd-nspawn", bash.Args...)
		}
	}
	err := RunCommand(bash, log.Info, os.Stderr)
	if err != nil {
		return err
	}

	PrintSuccess()

	return nil
}

func runParts(info forgeInfo) error {
	type action struct {
		part string
		args string
		do   bool
	}

	parts := []action{
		action{"configure", "./configure --prefix=/usr/", true},
		action{"build", "make", true},
		action{"test", "make test", info.test},
		action{"installpkg", "make DESTDIR=${dest_dir} install", true},
	}

	env := map[string]string{
		"MAKEFLAGS":              "-j6",
		"dest_dir":               info.workdir + dest,
		"FORCE_UNSAFE_CONFIGURE": "1", //TODO probably shouldn't do this
	}

	for _, part := range parts {
		if part.do {
			err := runPart(part.part, part.args, info, env)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func StripPackage(info forgeInfo) error {
	Header("Strip package")

	Clean := func(filter, strip string) error {
		cmd := fmt.Sprintf(`
			files=$(find %s -type f | grep %s)
			if ! [ -z "$files" ]; then
				strip %s $files
			fi
			`, info.root+info.workdir, filter, strip)

		return RunCommand(exec.Command("bash", "-c", cmd), log.Debug, os.Stderr)
	}
	Clean("/bin/", "-s")
	Clean("/sbin/", "-s")
	Clean("\\.so", "-s")
	Clean("\\.a$", "--strip-debug")

	PrintSuccess()
	return nil
}

func createSums(destdir string) (HashList, error) {
	hl := make(HashList)

	walkFunc := func(path string, f os.FileInfo, err error) (erri error) {
		if !f.IsDir() {
			sum, erri := Md5sum(path)
			if erri == nil {
				log.Debug.Format("%s:\t%s", sum, path)
				hl[path] = sum
			}
		}
		return
	}
	var err error
	InDir(destdir, func() {
		err = filepath.Walk(".", walkFunc)
	})
	return hl, err
}

func createPkgInstall(template string) (string, error) {
	buf := new(bytes.Buffer)
	bashStr := fmt.Sprintf(`
source %s
declare -f pre_install
declare -f post_install
exit 0
`, template)
	err := RunCommand(exec.Command("bash", "-c", bashStr), buf, os.Stderr)
	return buf.String(), err
}

func addFsToSpakg(dir, outfile string, archive spakg.Spakg) error {
	fsTarName := spakg.FsName
	fsTar := dir + "/" + fsTarName
	log.Debug.Println("Creating fs.tar: " + fsTar)

	var err error
	InDir(dir+dest, func() {
		err = RunCommand(exec.Command("tar", "-cvf", fsTar, "."), log.Debug, os.Stderr)
	})
	if err != nil {
		return err
	}
	log.Debug.Println()

	//Spakg
	log.Debug.Format("Creating package: %s", outfile)

	var innererr error
	err = WithFileReader(fsTar, func(fs io.Reader) {
		ie := WithFileWriter(outfile, true, func(tar io.Writer) {
			iie := archive.ToWriter(tar, fs)
			if iie != nil {
				innererr = iie
			}
		})
		if ie != nil {
			innererr = ie
		}
	})
	if err != nil {
		return err
	}
	if innererr != nil {
		return innererr
	}
	return nil
}

func BuildPackage(info forgeInfo) error {
	Header("Building package")

	//Md5Sums
	hl, err := createSums(info.root + info.workdir + dest)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to generate md5sums: %s", err))
	}

	pi := pkginfo.FromControl(&info.control)
	pi.BuildDate = time.Now()
	pi.SetFlagStates(info.states)

	//Template
	var templateStr string
	err = WithFileReader(info.template, func(reader io.Reader) {
		templateStr = ReaderToString(reader)
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to read template: %s", err))
	}

	//PkgInstall
	pkginstall, err := createPkgInstall(info.template)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to create pkginstall: %s", err))
	}

	//Create Spakg
	archive := spakg.Spakg{Md5sums: hl, Control: info.control, Template: templateStr, Pkginfo: *pi, Pkginstall: pkginstall}
	//FS
	err = addFsToSpakg(info.root+info.workdir, info.outfile, archive)
	if err != nil {
		return err
	}

	PrintSuccess()

	return nil
}
