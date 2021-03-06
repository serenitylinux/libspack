package wield

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/hash"
	"github.com/serenitylinux/libspack/repo"
	"github.com/serenitylinux/libspack/spakg"
)
import . "github.com/serenitylinux/libspack/misc"

func Wield(file string, destdir string) error {
	spkg, err := spakg.FromFile(file, nil)
	if err != nil {
		return err
	}

	err = PreInstall(spkg, destdir)
	if err != nil {
		return err
	}

	err = ExtractCheckCopy(file, destdir)
	if err != nil {
		return err
	}

	//Don't care if this fails
	Ldconfig(destdir)

	err = PostInstall(spkg, destdir)
	if err != nil {
		return err
	}

	return nil
}

func Ldconfig(destdir string) error {
	return RunCommand(exec.Command("ldconfig", "-r", destdir), log.Debug, os.Stderr)
}

func hasPart(part string, spkg *spakg.Spakg) bool {
	cmd := `
		%[1]s
		
		declare -f %[2]s > /dev/null
`
	cmd = fmt.Sprintf(cmd, spkg.Pkginstall, part)
	err := RunCommand(exec.Command("bash", "-c", cmd), log.Debug, os.Stderr)

	return err == nil
}

func runPart(part string, spkg *spakg.Spakg, destdir string) error {
	cmd := `
		%[1]s
		if ! [ -d /dev/ ]; then
			mkdir /dev;
		fi
		
		%[2]s
`
	cmd = fmt.Sprintf(cmd, spkg.Pkginstall, part)

	bash := exec.Command("bash", "-c", cmd)
	if filepath.Clean(destdir) != "/" {
		if _, err := exec.LookPath("chroot"); err == nil {
			bash.Args = append([]string{destdir}, bash.Args...)
			bash = exec.Command("chroot", bash.Args...)
		} else if _, err := exec.LookPath("systemd-nspawn"); err == nil {
			bash.Args = append([]string{"-D", destdir}, bash.Args...)
			bash = exec.Command("systemd-nspawn", bash.Args...)
		}
	}
	return RunCommand(bash, log.Debug, os.Stderr)
}

func PreInstall(pkg *spakg.Spakg, destdir string) error {
	if hasPart("pre_install", pkg) {
		HeaderFormat("PreInstall %s", pkg.Control.Name)
		err := runPart("pre_install", pkg, destdir)
		if err != nil {
			return err
		}
		PrintSuccess()
	}
	return nil
}
func PostInstall(pkg *spakg.Spakg, destdir string) error {
	if hasPart("post_install", pkg) {
		HeaderFormat("PostInstall %s", pkg.Control.Name)
		err := runPart("post_install", pkg, destdir)
		if err != nil {
			return err
		}
		PrintSuccess()
	}
	return nil
}

func ExtractCheckCopy(pkgfile string, destdir string) error {

	tmpDir, _ := ioutil.TempDir(os.TempDir(), "wield")
	defer os.RemoveAll(tmpDir)

	pkg, err := spakg.FromFile(pkgfile, &tmpDir)
	if err != nil {
		return err
	}

	fsDir := tmpDir + "/fs"
	if err = os.MkdirAll(fsDir, 0755); err != nil {
		return err
	}

	HeaderFormat("Unpacking  %s", pkg.Control.Name)
	cmd := exec.Command("tar", "-xvpf", tmpDir+"/fs.tar", "-C", fsDir)
	err = RunCommand(cmd, log.Debug, os.Stderr)
	if err != nil {
		return err
	}

	PrintSuccess()

	HeaderFormat("Checking   %s", pkg.Control.Name)

	walk := func(path string, f os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if !f.IsDir() && !IsSymlink(f) {
			origSum, exists := pkg.Md5sums[path]
			if !exists {
				return errors.New(fmt.Sprintf("Sum for %s does not exist", path))
			}

			sum, erri := hash.Md5sum(path)
			if erri != nil {
				return errors.New(fmt.Sprintf("Cannot compute sum of %s", path))
			}

			if origSum != sum {
				return errors.New(fmt.Sprintf("Sum of %s does not match. Expected %s, calculated %s", path, origSum, sum))
			}
			log.Debug.Format("%s\t: %s", sum, path)
		}
		return nil
	}

	InDir(fsDir, func() {
		err = filepath.Walk(".", walk)
	})

	if err != nil {
		return err
	}
	PrintSuccess()

	HeaderFormat("Installing %s", pkg.Control.Name)

	copyWalk := func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fsPath := fsDir + "/" + path
		destPath := destdir + path

		if IsSymlink(f) {
			target, e := os.Readlink(fsPath)
			if e != nil {
				return e
			}

			//Let's just wing it!
			os.Remove(destPath)

			e = os.Symlink(target, destPath)
			if e != nil {
				return e
			}
		} else if f.IsDir() {
			if !PathExists(destPath) {
				e := os.MkdirAll(destPath, f.Mode())
				if e != nil {
					return e
				}
			}
		} else {
			var currhash string
			var e error
			if PathExists(destPath) {
				currhash, e = hash.Md5sum(destPath)
				if e != nil {
					return e
				}
			}
			if currhash != pkg.Md5sums[path] {
				if PathExists(destPath) {
					log.Debug.Format("Removing current %v", destPath)
					e = os.Remove(destPath)
					if e != nil {
						return e
					}
				}
				log.Debug.Format("Creating %v", destPath)
				e = CopyFile(fsPath, destPath)
				if e != nil {
					return e
				}
			} else {
				log.Debug.Format("Skipping unchanged %s", path)
			}
		}

		uid, gid := GetUidGid(f)
		os.Lchown(destPath, uid, gid)
		os.Chmod(destPath, f.Mode())
		return nil
		//TODO collisions and changed conf files
	}
	InDir(fsDir, func() {
		err = filepath.Walk(".", copyWalk)
	})
	if err != nil {
		return err
	}

	if prev, _ := repo.GetPackageInstalledByName(pkg.Control.Name, destdir); prev != nil {
		log.Debug.Format("Removing files from old version %s", prev.PkgInfo.PrettyString())
		//TODO remove empty leftover dirs
		for oldf, _ := range prev.Hashes {
			_, skip := pkg.Md5sums[oldf]
			if !skip {
				log.Debug.Format("Removing %s", destdir+oldf)
				err = os.RemoveAll(destdir + oldf)
				if err != nil {
					log.Warn.Format("Could not remove %s from old version, %v", destdir+oldf, err)
				}
			} else {
				log.Debug.Format("Keeping %s", destdir+oldf)
			}
		}
	}

	Ldconfig(destdir)

	PrintSuccess()
	return nil
}
