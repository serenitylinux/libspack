package spakg

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/hash"
	"github.com/serenitylinux/libspack/misc"
	"github.com/serenitylinux/libspack/pkginfo"
)

const (
	ControlName    = "pkg.control"
	PkginfoName    = "pkginfo.txt"
	TemplateName   = "pkg.template"
	Md5sumsName    = "md5sums.txt"
	PkgInstallName = "pkginstall.sh"
	FsName         = "fs.tar"
)

type Spakg struct {
	Pkginfo    pkginfo.PkgInfo
	Control    control.Control
	Md5sums    hash.HashList
	Pkginstall string
	Template   string
}

func (s *Spakg) ToFile(filename string, fsReader io.Reader) (err error) {
	writerFunc := func(w io.Writer) { err = s.ToWriter(w, fsReader) }

	ioerr := misc.WithFileWriter(filename, true, writerFunc)
	if ioerr != nil {
		return ioerr
	}

	return
}

func writeTarBytes(tw *tar.Writer, name string, bytes []byte, length int64) error {
	hdr := &tar.Header{
		Name:    name,
		ModTime: time.Now(),
		Size:    length,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	_, err := tw.Write(bytes)
	return err
}

func writeTarEntry(tw *tar.Writer, name string, reader io.Reader) error {
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(reader)
	if err != nil {
		return err
	}
	return writeTarBytes(tw, name, buf.Bytes(), n)
}

func writeTarString(tw *tar.Writer, name string, val string) error {
	bytes := []byte(val)
	return writeTarBytes(tw, name, bytes, int64(len(bytes)))
}
func writeTarJSON(tw *tar.Writer, name string, val interface{}) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return writeTarBytes(tw, name, bytes, int64(len(bytes)))
}

func (s *Spakg) ToWriter(writer io.Writer, fsReader io.Reader) (err error) {
	tw := tar.NewWriter(writer)

	err = writeTarJSON(tw, ControlName, s.Control)
	if err != nil {
		return
	}
	err = writeTarJSON(tw, PkginfoName, s.Pkginfo)
	if err != nil {
		return
	}
	err = writeTarString(tw, TemplateName, s.Template)
	if err != nil {
		return
	}
	err = writeTarJSON(tw, Md5sumsName, s.Md5sums)
	if err != nil {
		return
	}
	err = writeTarString(tw, PkgInstallName, s.Pkginstall)
	if err != nil {
		return
	}
	err = writeTarEntry(tw, FsName, fsReader)
	if err != nil {
		return
	}

	return err
}

func FromFile(filename string, tarname *string) (s *Spakg, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return FromReader(file, tarname)
}

func FromReader(reader io.Reader, tarname *string) (*Spakg, error) {
	var s Spakg
	tr := tar.NewReader(reader)
	decoder := json.NewDecoder(tr)
	foundControl := false
	foundPkginfo := false
	foundFs := false
	//	foundTemplate := false
	foundPkginstall := false
	foundmd5sum := false

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch hdr.Name {
		case ControlName:
			err = decoder.Decode(&s.Control)
			if err != nil {
				return nil, err
			}
			foundControl = true
		case PkginfoName:
			err = decoder.Decode(&s.Pkginfo)
			if err != nil {
				return nil, err
			}
			foundPkginfo = true
		case TemplateName:
			s.Template = misc.ReaderToString(tr)
			//				foundTemplate = true
		case PkgInstallName:
			s.Pkginstall = misc.ReaderToString(tr)
			foundPkginstall = true
		case Md5sumsName:
			err = decoder.Decode(&s.Md5sums)
			if err != nil {
				return nil, err
			}
			foundmd5sum = true
		case FsName:
			if tarname != nil {
				err := misc.WithFileWriter(*tarname+"/"+FsName, true, func(fsw io.Writer) {
					io.Copy(fsw, tr)
				})
				if err != nil {
					return nil, err
				}
			}
			foundFs = true
		default:
			return nil, errors.New(fmt.Sprintf("Invalid Spakg, contains %s", hdr.Name))
		}
	}
	//Template may not be nessesary
	if foundControl && foundPkginfo && foundFs && foundPkginstall && foundmd5sum {
		return &s, nil
	} else {
		//TODO what file is missing
		return nil, errors.New("Invalid Spakg, missing files")
	}
}
