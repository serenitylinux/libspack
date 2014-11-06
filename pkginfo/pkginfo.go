package pkginfo

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"time"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/flag"
	"github.com/serenitylinux/libspack/helpers/json"
)

type PkgInfo struct {
	Name             string
	Version          string
	Iteration        int
	BuildDate        time.Time
	Flags            []string
	parsedFlags      flag.FlagSetList
	FlagStates       []string
	parsedFlagStates flag.FlagList
}

type PkgInfoList []PkgInfo

func (p *PkgInfo) String() string {
	return json.Stringify(p)
}
func (p *PkgInfo) UUID() string {
	return fmt.Sprintf("%s-%s_%d_%x", p.Name, p.Version, p.Iteration, p.flagHash())
}
func (p *PkgInfo) PrettyString() string {
	return fmt.Sprintf("%s %s_%d (%s)", p.Name, p.Version, p.Iteration, p.ComputedFlagStates())
}

func (p *PkgInfo) flagHash() uint32 {
	str := p.Name
	for _, flag := range p.ComputedFlagStates() {
		str += flag.String()
	}
	return crc32.ChecksumIEEE([]byte(str))
}

func FromControl(c *control.Control) *PkgInfo {
	p := PkgInfo{Name: c.Name, Version: c.Version, Flags: c.Flags, Iteration: c.Iteration}
	return &p
}

func (p *PkgInfo) InstanceOf(c *control.Control) bool {
	return c.Name == p.Name && p.Version == c.Version && c.Iteration == p.Iteration
}

func (p *PkgInfo) ToFile(filename string) error {
	return json.EncodeFile(filename, true, p)
}

func (p *PkgInfo) ParsedFlags() flag.FlagSetList {
	if p.parsedFlags == nil {
		p.parsedFlags = make([]flag.FlagSet, 0)
		for _, s := range p.Flags {
			flag, err := flag.FromString(s)
			if err != nil {
				log.Warn.Format("Invalid flag in package %s '%s': %s", p.Name, s, err)
				continue
			}
			p.parsedFlags = append(p.parsedFlags, flag)
		}
	}
	return p.parsedFlags
}

func (p *PkgInfo) ParsedFlagStates() flag.FlagList {
	if p.parsedFlagStates == nil {
		p.parsedFlagStates = make(flag.FlagList, 0)
		for _, s := range p.FlagStates {
			flag, err := flag.FlagFromString(s)
			if err != nil {
				log.Warn.Format("Invalid flag in package %s '%s': %s", p.Name, s, err)
				continue
			}
			p.parsedFlagStates = append(p.parsedFlagStates, *flag)
		}
	}
	return p.parsedFlagStates
}
func (p *PkgInfo) SetFlagState(f *flag.Flag) error {
	p.parsedFlagStates = nil

	if !p.ParsedFlags().Contains(f.Name) {
		return errors.New("Invalid flag " + f.Name)
	}

	for i, v := range p.FlagStates {
		if v[1:] == f.Name { //equals ignore sign
			p.FlagStates[i] = f.String()
			return nil
		}
	}
	//Does not contain f already
	p.FlagStates = append(p.FlagStates, f.String())
	return nil
}

func (p *PkgInfo) SetFlagStates(states []flag.Flag) error {
	for _, f := range states {
		err := p.SetFlagState(&f)
		if err != nil {
			return err
		}
	}
	return nil
}

//Default + Configured
func (p *PkgInfo) ComputedFlagStates() flag.FlagList {
	res := make(flag.FlagList, 0)
	for _, f := range p.ParsedFlags() {
		res = append(res, f.Flag)
	}

	for _, parsedf := range p.ParsedFlagStates() {
		for i, currf := range res {
			if currf.Name == parsedf.Name {
				res[i] = parsedf
				break
			}
		}
	}

	return res
}

func (p *PkgInfo) Satisfies(flags flag.FlagList) bool {
	return flags.IsSubSet(p.ComputedFlagStates())
}

func FromFile(filename string) (*PkgInfo, error) {
	var i PkgInfo
	err := json.DecodeFile(filename, &i)
	return &i, err
}

func FromReader(reader io.Reader) (*PkgInfo, error) {
	var i PkgInfo
	err := json.DecodeReader(reader, &i)
	return &i, err
}
