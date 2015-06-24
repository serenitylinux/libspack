package crunch

import (
	"encoding/json"
	"testing"

	"github.com/cam72cam/go-lumberjack/log"
	"github.com/serenitylinux/libspack/repo"
	"github.com/serenitylinux/libspack/spdl"
)

func TestCrunchBasics(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	loadEntry := func(c string) (e repo.Entry) {
		err := json.Unmarshal([]byte(c), &e.Control)
		if err != nil {
			panic(err)
		}
		e.Template = "does_not_exist_" + e.Control.Name + ".pie"
		return e
	}
	A := loadEntry(`
{
	"Name": "A",
	"Description": "A's Description",
	"Version": "1.0.0",
	"Iteration": 1,
	"Deps": [
		"[-dev]B",
		"[+dev]C"
	],
	"Flags": [
		"-dev"
	]
}`)
	B := loadEntry(`
{
	"Name": "B",
	"Description": "B's Description",
	"Version": "1.0.0",
	"Iteration": 1,
	"Deps": [
		"C"
	]
}`)

	C := loadEntry(`
{
	"Name": "C",
	"Description": "C's Description",
	"Version": "1.0.0",
	"Iteration": 1,
	"Deps": [
		"A(+dev)"
	]
}`)

	g, err := NewGraph("/test_dir_does_not_exist", repo.RepoList{"Test": repo.MockRepo("Test", A, B, C)})
	if err != nil {
		t.Error(err)
	}
	g.EnablePackage(spdl.Dep{Name: "A"}, InstallConvenient)

	err = g.Crunch()
	if err != nil {
		t.Error(err)
	}

	ANode := g.nodes["A"]
	if !ANode.IsEnabled() {
		t.Errorf("A should have been enabled")
	}
	BNode := g.nodes["B"]
	if BNode.IsEnabled() {
		t.Errorf("B should NOT have been enabled")
	}
	CNode := g.nodes["C"]
	if !CNode.IsEnabled() {
		t.Errorf("C should have been enabled")
	}

}
