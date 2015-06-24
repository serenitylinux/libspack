package spdl

import (
	"reflect"
	"testing"
)

func TestFlagListString(t *testing.T) {
	fl := buildFlagList(
		Flag{Name: "foo", State: Enabled},
		Flag{Name: "bar", State: Disabled},
		Flag{Name: "baz", State: Inherit},
		Flag{Name: "bash", State: Invert},
	)

	ffl := buildFlatFlagList(
		FlatFlag{Name: "foo", Enabled: true},
		FlatFlag{Name: "bar", Enabled: false},
	)
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
		sub    FlatFlagList
		super  FlatFlagList
		expect bool
	}
	cases := []Case{
		{
			name:   "Empty",
			expect: true,
		},
		{
			name: "Equal",
			sub: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
			),
			super: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
			),
			expect: true,
		},
		{
			name: "Subset",
			sub: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
			),
			super: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
			),
			expect: true,
		},
		{
			name: "Super",
			sub: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
			),
			super: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
			),
			expect: false,
		},
		{
			name: "Diff enabled",
			sub: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: false},
				FlatFlag{Name: "bar", Enabled: false},
			),
			super: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
			),
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
	ffl := buildFlatFlagList(
		FlatFlag{Name: "foo", Enabled: true},
		FlatFlag{Name: "bar", Enabled: false},
	)
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
	orig := buildFlagList(
		Flag{Name: "foo", State: Enabled},
		Flag{Name: "bar", State: Disabled},
		Flag{Name: "baz", State: Inherit},
		Flag{Name: "bash", State: Invert},
	)
	clone := orig.Clone()
	if !reflect.DeepEqual(orig, clone) {
		t.Errorf("Clones are not equal to the original")
	}

	clone.Add(Flag{Name: "foo", State: Disabled})
	if f, ok := orig.Contains("foo"); !ok || f.State != Enabled {
		t.Errorf("Not a deep enough clone")
	}
}

func TestFlagListDefaults(t *testing.T) {
	type Case struct {
		name     string
		original FlagList
		defaults FlatFlagList
		expect   FlatFlagList
		err      bool
	}
	cases := []Case{
		{
			name: "Simple (no defaults)",
			original: buildFlagList(
				Flag{Name: "foo", State: Enabled},
				Flag{Name: "bar", State: Disabled},
			),
			expect: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
			),
		},
		{
			name: "With Defaults",
			original: buildFlagList(
				Flag{Name: "foo", State: Enabled},
				Flag{Name: "bar", State: Disabled},
				Flag{Name: "baz", State: Inherit},
				Flag{Name: "basha", State: Invert},
				Flag{Name: "bashb", State: Invert},
			),
			defaults: buildFlatFlagList(
				FlatFlag{Name: "baz", Enabled: true},
				FlatFlag{Name: "basha", Enabled: true},
				FlatFlag{Name: "bashb", Enabled: false},
			),
			expect: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
				FlatFlag{Name: "baz", Enabled: true},
				FlatFlag{Name: "basha", Enabled: false},
				FlatFlag{Name: "bashb", Enabled: true},
			),
		},
		{
			name: "Missing",
			original: buildFlagList(
				Flag{Name: "foo", State: Enabled},
				Flag{Name: "bar", State: Disabled},
				Flag{Name: "baz", State: Inherit},
				Flag{Name: "basha", State: Invert},
				Flag{Name: "bashb", State: Invert},
			),
			defaults: buildFlatFlagList(
				FlatFlag{Name: "baz", Enabled: true},
				FlatFlag{Name: "basha", Enabled: true},
			),
			expect: buildFlatFlagList(
				FlatFlag{Name: "foo", Enabled: true},
				FlatFlag{Name: "bar", Enabled: false},
				FlatFlag{Name: "baz", Enabled: true},
				FlatFlag{Name: "basha", Enabled: false},
				FlatFlag{Name: "bashb", Enabled: true},
			),
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
