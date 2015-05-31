package flag_test

import (
	"reflect"
	"testing"

	"github.com/serenitylinux/libspack/flag"
)

func TestFlagListString(t *testing.T) {
	fl := flag.FlagList{
		"foo":  flag.Flag{Name: "foo", State: flag.Enabled},
		"bar":  flag.Flag{Name: "bar", State: flag.Disabled},
		"baz":  flag.Flag{Name: "baz", State: flag.Inherit},
		"bash": flag.Flag{Name: "bash", State: flag.Invert},
	}

	ffl := flag.FlatFlagList{
		"foo": flag.FlatFlag{Name: "foo", Enabled: true},
		"bar": flag.FlatFlag{Name: "bar", Enabled: false},
	}
	if expect, actual := "+foo -bar ?baz ~bash", fl.String(); expect != actual {
		t.Errorf("Expected %s, got %s", expect, actual)
	}
	if expect, actual := "+foo -bar", ffl.String(); expect != actual {
		t.Errorf("Expected %s, got %s", expect, actual)
	}
	fl.ColorString()
	ffl.ColorString()
}

func TestFlagListSubset(t *testing.T) {
	type Case struct {
		name   string
		sub    flag.FlatFlagList
		super  flag.FlatFlagList
		expect bool
	}
	cases := []Case{
		{
			name:   "Empty",
			expect: true,
		},
		{
			name: "Equal",
			sub: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
				"bar": flag.FlatFlag{Name: "bar", Enabled: false},
			},
			super: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
				"bar": flag.FlatFlag{Name: "bar", Enabled: false},
			},
			expect: true,
		},
		{
			name: "Subset",
			sub: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
			},
			super: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
				"bar": flag.FlatFlag{Name: "bar", Enabled: false},
			},
			expect: true,
		},
		{
			name: "Super",
			sub: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
				"bar": flag.FlatFlag{Name: "bar", Enabled: false},
			},
			super: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
			},
			expect: false,
		},
		{
			name: "Diff enabled",
			sub: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: false},
				"bar": flag.FlatFlag{Name: "bar", Enabled: false},
			},
			super: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
				"bar": flag.FlatFlag{Name: "bar", Enabled: false},
			},
			expect: false,
		},
	}
	for _, c := range cases {
		t.Logf(c.name)
		actual := c.sub.IsSubsetOf(c.super)
		if actual != c.expect {
			tobe := "to be"
			if !c.expect {
				tobe = "not to be"
			} //what was the question?
			t.Errorf("Expected %s %s a subset of %s", c.sub, tobe, c.super)
		}
	}
}

func TestFlagListEnabled(t *testing.T) {
	ffl := flag.FlatFlagList{
		"foo": flag.FlatFlag{Name: "foo", Enabled: true},
		"bar": flag.FlatFlag{Name: "bar", Enabled: false},
	}
	type Case struct {
		name   string
		flag   string
		expect bool
	}
	cases := []Case{
		{
			name:   "Ok (true)",
			flag:   "foo",
			expect: true,
		},
		{
			name:   "Ok (false)",
			flag:   "bar",
			expect: false,
		},
		{
			name:   "Missing",
			flag:   "wablahflababablah",
			expect: false,
		},
	}
	for _, c := range cases {
		t.Log(c.name)
		actual := ffl.IsEnabled(c.flag)
		if actual != c.expect {
			t.Errorf("Expected %s, got %s", c.expect, actual)
		}
	}
}

func TestFlagListClone(t *testing.T) {
	orig := flag.FlagList{
		"foo":  flag.Flag{Name: "foo", State: flag.Enabled},
		"bar":  flag.Flag{Name: "bar", State: flag.Disabled},
		"baz":  flag.Flag{Name: "baz", State: flag.Inherit},
		"bash": flag.Flag{Name: "bash", State: flag.Invert},
	}
	clone := orig.Clone()
	if !reflect.DeepEqual(orig, clone) {
		t.Errorf("Clones are not equal to the original")
	}

	foo := clone["foo"]
	foo.State = flag.Disabled

	clone["foo"] = foo
	if orig["foo"].State != flag.Enabled {
		t.Errorf("Not a deep enough clone")
	}
}

func TestFlagListDefaults(t *testing.T) {
	type Case struct {
		name     string
		original flag.FlagList
		defaults flag.FlatFlagList
		expect   flag.FlatFlagList
		err      bool
	}
	cases := []Case{
		{
			name: "Simple (no defaults)",
			original: flag.FlagList{
				"foo": flag.Flag{Name: "foo", State: flag.Enabled},
				"bar": flag.Flag{Name: "bar", State: flag.Disabled},
			},
			expect: flag.FlatFlagList{
				"foo": flag.FlatFlag{Name: "foo", Enabled: true},
				"bar": flag.FlatFlag{Name: "bar", Enabled: false},
			},
		},
		{
			name: "With Defaults",
			original: flag.FlagList{
				"foo":   flag.Flag{Name: "foo", State: flag.Enabled},
				"bar":   flag.Flag{Name: "bar", State: flag.Disabled},
				"baz":   flag.Flag{Name: "baz", State: flag.Inherit},
				"basha": flag.Flag{Name: "basha", State: flag.Invert},
				"bashb": flag.Flag{Name: "bashb", State: flag.Invert},
			},
			defaults: flag.FlatFlagList{
				"baz":   flag.FlatFlag{Name: "baz", Enabled: true},
				"basha": flag.FlatFlag{Name: "basha", Enabled: true},
				"bashb": flag.FlatFlag{Name: "bashb", Enabled: false},
			},
			expect: flag.FlatFlagList{
				"foo":   flag.FlatFlag{Name: "foo", Enabled: true},
				"bar":   flag.FlatFlag{Name: "bar", Enabled: false},
				"baz":   flag.FlatFlag{Name: "baz", Enabled: true},
				"basha": flag.FlatFlag{Name: "basha", Enabled: false},
				"bashb": flag.FlatFlag{Name: "bashb", Enabled: true},
			},
		},
		{
			name: "Missing",
			original: flag.FlagList{
				"foo":   flag.Flag{Name: "foo", State: flag.Enabled},
				"bar":   flag.Flag{Name: "bar", State: flag.Disabled},
				"baz":   flag.Flag{Name: "baz", State: flag.Inherit},
				"basha": flag.Flag{Name: "basha", State: flag.Invert},
				"bashb": flag.Flag{Name: "bashb", State: flag.Invert},
			},
			defaults: flag.FlatFlagList{
				"baz":   flag.FlatFlag{Name: "baz", Enabled: true},
				"basha": flag.FlatFlag{Name: "basha", Enabled: true},
			},
			expect: flag.FlatFlagList{
				"foo":   flag.FlatFlag{Name: "foo", Enabled: true},
				"bar":   flag.FlatFlag{Name: "bar", Enabled: false},
				"baz":   flag.FlatFlag{Name: "baz", Enabled: true},
				"basha": flag.FlatFlag{Name: "basha", Enabled: false},
				"bashb": flag.FlatFlag{Name: "bashb", Enabled: true},
			},
			err: true,
		},
	}
	for _, c := range cases {
		t.Log(c.name)
		actual, err := c.original.WithDefaults(c.defaults)
		if c.err {
			if err == nil {
				t.Error("Expected error")
			} else {
				t.Logf("Ok")
			}
			continue
		}
		if !reflect.DeepEqual(c.expect, actual) {
			t.Errorf("Expected %v, got %v", c.expect, actual)
		}
		t.Log("Ok")
	}
}
