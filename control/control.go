package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/serenitylinux/libspack/dep"
	"github.com/serenitylinux/libspack/flag/expr"
)

type Control struct {
	Name        string
	Version     string
	Iteration   int
	Description string
	Url         string
	Src         []string
	Arch        []string

	Bdeps dep.DepList
	Deps  dep.DepList
	Flags expr.FlagSetList
	//Provides (libjpeg, cc)
	//Provides Hook (update mime types)
}

func (c *Control) String() string {
	return fmt.Sprintf("%s-%s_%d", c.Name, c.Version, c.Iteration)
}

func FromTemplateFile(template string) (c Control, err error) {
	commands := `
template=` + template + `
default=$(basedir $template)/default

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
},
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
