package constraintconfig

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/misc"
	"github.com/serenitylinux/libspack/spdl"
)

type ConstraintList map[string]spdl.Dep

func (list ConstraintList) addFile(path string) error {
	var interr error
	err := misc.WithFileReader(path, func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				continue
			}

			d, err := spdl.ParseDep(line)

			if err != nil {
				interr = err
				return
			}

			if d.Condition != nil {
				interr = errors.New("Cannot have a condition in a constraint config file: " + line)
				return
			}
			if d.Version1 == nil && d.Version2 == nil && (d.Flags == nil || len(d.Flags.Slice()) == 0) {
				interr = errors.New("Package " + d.Name + " has no constraints specified")
				return
			}

			list[d.Name] = d
		}
		if err := scanner.Err(); err != nil {
			interr = err
		}
	})

	if interr != nil {
		return interr
	}
	return err
}

var cached = make(map[string]ConstraintList)

func GetAll(root string) ConstraintList {
	if list, exists := cached[root]; exists {
		return list
	}

	pre := root + "/etc/spack/pkg"
	fl := make(ConstraintList, 0)

	if misc.PathExists(pre + ".conf") {
		err := fl.addFile(pre + ".conf")
		if err != nil {
			log.Error.Println(err)
			return nil
		}
	}

	if misc.PathExists(pre) {
		err := filepath.Walk(pre, func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() {
				return fl.addFile(path)
			}
			return nil
		})
		if err != nil {
			log.Error.Println(err)
			return nil
		}
	}

	cached[root] = fl
	return fl
}
