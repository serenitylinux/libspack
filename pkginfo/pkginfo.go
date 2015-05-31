package pkginfo

import (
	"fmt"
	"hash/crc32"
	"time"

	"github.com/serenitylinux/libspack/control"
	"github.com/serenitylinux/libspack/flag"
)

type PkgInfo struct {
	Name       string
	Version    string
	Iteration  int
	BuildDate  time.Time
	FlagStates flag.FlatFlagList
}

type PkgInfoList []PkgInfo

func (p *PkgInfo) String() string {
	return fmt.Sprintf("%s-%s_%d_%x", p.Name, p.Version, p.Iteration, p.flagHash())
}
func (p *PkgInfo) PrettyString() string {
	return fmt.Sprintf("%s %s_%d (%s)", p.Name, p.Version, p.Iteration, p.FlagStates)
}

func (p *PkgInfo) flagHash() uint32 {
	str := p.Name
	for _, flag := range p.FlagStates {
		str += flag.String()
	}
	return crc32.ChecksumIEEE([]byte(str))
}

func FromControl(c *control.Control) *PkgInfo {
	p := PkgInfo{
		Name:       c.Name,
		Version:    c.Version,
		FlagStates: c.Flags.Defaults(),
		Iteration:  c.Iteration,
	}
	return &p
}

func (p *PkgInfo) InstanceOf(c *control.Control) bool {
	return c.Name == p.Name && p.Version == c.Version && c.Iteration == p.Iteration
}

func (p *PkgInfo) SetFlagState(f flag.FlatFlag) {
	p.FlagStates[f.Name] = f
}

func (p *PkgInfo) SetFlagStates(states []flag.FlatFlag) {
	for _, f := range states {
		p.SetFlagState(f)
	}
}

func (p *PkgInfo) Satisfies(flags flag.FlatFlagList) bool {
	return flags.IsSubsetOf(p.FlagStates)
}
