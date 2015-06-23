package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/serenitylinux/libspack/spdl"
)

type Control struct {
	Name        string
	Version     string
	Iteration   int
	Description string
	Url         string
	Src         []string
	Arch        []string

	Bdeps spdl.DepList
	Deps  spdl.DepList
	Flags spdl.FlagExprList
	//Provides (libjpeg, cc)
	//Provides Hook (update mime types)
}

//Hack for older controls for now
func (c *Control) FlagsCrunch() {
	//changed := false
	crunch := func(oldl spdl.DepList) (newl spdl.DepList) {
		for _, o := range oldl {
			dupeIndex := -1
			for i, n := range newl {
				if o.Name == n.Name {
					//fmt.Printf("CRUNCH: %v %v:%v\n", c.Name, o.String(), n.String())
					dupeIndex = i
					break
				}
			}
			if dupeIndex == -1 {
				newl = append(newl, o)
			} else {
				//changed = true
				o.Condition = nil
				if o.Flags != nil && len(o.Flags.Slice()) != 0 {
					o.Flags.Add(spdl.Flag{Name: o.Flags.Slice()[0].Name, State: spdl.Inherit})
				}
				newl[dupeIndex] = o
			}
		}
		return newl
	}
	c.Bdeps = crunch(c.Bdeps)
	c.Deps = crunch(c.Deps)
	/*
		if changed {
			s, _ := json.MarshalIndent(c, "", "\t")
			fmt.Println(string(s))
		}*/
}

func (c Control) String() string {
	return fmt.Sprintf("%s-%s_%d", c.Name, c.Version, c.Iteration)
}
func (c Control) Equals(other Control) bool {
	//TODO better comparison?
	return c.String() == other.String()
}

func FromTemplateFile(template string) (c Control, err error) {
	commands := `
template=` + template + `
default=$(dirname $template)/default

. $template
if [ -f "$default" ]; then
	. $default
fi

function lister() {
	local set i
	set=""
	for i in "$@"; do
		echo -en "$set\"$i\""
		set=", "
	done
}

srcval="$(lister ${src[@]})"
bdepsval="$(lister ${bdeps[@]})"
depsval="$(lister ${deps[@]})"
archval="$(lister ${arch[@]})"
flagsval="$(lister ${flags[@]})"

cat << EOT
{
  "Name": "$name",
  "Version": "$version",
  "Iteration": $iteration,
  "Description": "$desc",
  "Url": "$url",
  "Src": [$srcval],
  "Bdeps": [ $bdepsval ],
  "Deps": [ $depsval ],
  "Arch": [ $archval ],
  "Flags": [ $flagsval ]
}
EOT`
	var buf bytes.Buffer
	cmd := exec.Command("bash", "-ec", commands)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(buf.Bytes(), &c)
	return c, err
}
