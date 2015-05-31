package spdl_test

import (
	"reflect"
	"testing"

	"github.com/serenitylinux/libspack/spdl"
)

func TestFlagListString(t *testing.T) {
	fl := spdl.FlagList{
		"foo":  spdl.Flag{Name: "foo", State: spdl.Enabled},
		"bar":  spdl.Flag{Name: "bar", State: spdl.Disabled},
		"baz":  spdl.Flag{Name: "baz", State: spdl.Inherit},
		"bash": spdl.Flag{Name: "bash", State: spdl.Invert},
	}

	ffl := spdl.FlatFlagList{
		"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
		"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
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
		sub    spdl.FlatFlagList
		super  spdl.FlatFlagList
		expect bool
	}
	cases := []Case{
		{
			name:   "Empty",
			expect: true,
		},
		{
			name: "Equal",
			sub: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
			},
			super: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
			},
			expect: true,
		},
		{
			name: "Subset",
			sub: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
			},
			super: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
			},
			expect: true,
		},
		{
			name: "Super",
			sub: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
			},
			super: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
			},
			expect: false,
		},
		{
			name: "Diff enabled",
			sub: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: false},
				"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
			},
			super: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
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
	ffl := spdl.FlatFlagList{
		"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
		"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
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
	orig := spdl.FlagList{
		"foo":  spdl.Flag{Name: "foo", State: spdl.Enabled},
		"bar":  spdl.Flag{Name: "bar", State: spdl.Disabled},
		"baz":  spdl.Flag{Name: "baz", State: spdl.Inherit},
		"bash": spdl.Flag{Name: "bash", State: spdl.Invert},
	}
	clone := orig.Clone()
	if !reflect.DeepEqual(orig, clone) {
		t.Errorf("Clones are not equal to the original")
	}

	foo := clone["foo"]
	foo.State = spdl.Disabled

	clone["foo"] = foo
	if orig["foo"].State != spdl.Enabled {
		t.Errorf("Not a deep enough clone")
	}
}

func TestFlagListDefaults(t *testing.T) {
	type Case struct {
		name     string
		original spdl.FlagList
		defaults spdl.FlatFlagList
		expect   spdl.FlatFlagList
		err      bool
	}
	cases := []Case{
		{
			name: "Simple (no defaults)",
			original: spdl.FlagList{
				"foo": spdl.Flag{Name: "foo", State: spdl.Enabled},
				"bar": spdl.Flag{Name: "bar", State: spdl.Disabled},
			},
			expect: spdl.FlatFlagList{
				"foo": spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar": spdl.FlatFlag{Name: "bar", Enabled: false},
			},
		},
		{
			name: "With Defaults",
			original: spdl.FlagList{
				"foo":   spdl.Flag{Name: "foo", State: spdl.Enabled},
				"bar":   spdl.Flag{Name: "bar", State: spdl.Disabled},
				"baz":   spdl.Flag{Name: "baz", State: spdl.Inherit},
				"basha": spdl.Flag{Name: "basha", State: spdl.Invert},
				"bashb": spdl.Flag{Name: "bashb", State: spdl.Invert},
			},
			defaults: spdl.FlatFlagList{
				"baz":   spdl.FlatFlag{Name: "baz", Enabled: true},
				"basha": spdl.FlatFlag{Name: "basha", Enabled: true},
				"bashb": spdl.FlatFlag{Name: "bashb", Enabled: false},
			},
			expect: spdl.FlatFlagList{
				"foo":   spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar":   spdl.FlatFlag{Name: "bar", Enabled: false},
				"baz":   spdl.FlatFlag{Name: "baz", Enabled: true},
				"basha": spdl.FlatFlag{Name: "basha", Enabled: false},
				"bashb": spdl.FlatFlag{Name: "bashb", Enabled: true},
			},
		},
		{
			name: "Missing",
			original: spdl.FlagList{
				"foo":   spdl.Flag{Name: "foo", State: spdl.Enabled},
				"bar":   spdl.Flag{Name: "bar", State: spdl.Disabled},
				"baz":   spdl.Flag{Name: "baz", State: spdl.Inherit},
				"basha": spdl.Flag{Name: "basha", State: spdl.Invert},
				"bashb": spdl.Flag{Name: "bashb", State: spdl.Invert},
			},
			defaults: spdl.FlatFlagList{
				"baz":   spdl.FlatFlag{Name: "baz", Enabled: true},
				"basha": spdl.FlatFlag{Name: "basha", Enabled: true},
			},
			expect: spdl.FlatFlagList{
				"foo":   spdl.FlatFlag{Name: "foo", Enabled: true},
				"bar":   spdl.FlatFlag{Name: "bar", Enabled: false},
				"baz":   spdl.FlatFlag{Name: "baz", Enabled: true},
				"basha": spdl.FlatFlag{Name: "basha", Enabled: false},
				"bashb": spdl.FlatFlag{Name: "bashb", Enabled: true},
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
