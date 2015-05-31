package spdl_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/serenitylinux/libspack/spdl"
)

/* Errata:
Ordering of the json list may change during the test
simply re-run the test a few times and it should pass
TODO: compare order-agnostic
*/

func TestFlagListEncoding(t *testing.T) {
	type Case struct {
		name string
		json string
		err  bool
	}
	cases := []Case{
		{
			name: "Empty",
			json: "[]",
		},
		{
			name: "Single Simple",
			json: `["-foo"]`,
		},
		{
			name: "Multiple",
			json: `["-foo","+bar","?baz","~bash"]`,
		},
	}
	for _, c := range cases {
		t.Logf(c.name)
		str := fmt.Sprintf(`{"Flags":%s}`, c.json)
		var res struct {
			Flags spdl.FlagList
		}
		err := json.Unmarshal([]byte(str), &res)
		if c.err {
			if err == nil {
				t.Fatalf("Expected a error, got nil")
			} else {
				t.Logf("Ok")
			}
			continue
		} else {
			if err != nil {
				t.Fatalf("Unexpected error %s", err.Error())
				continue
			}
		}

		bres, err := json.Marshal(&res)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
			continue
		}
		if str != string(bres) {
			t.Fatalf("Expected %s, got %s", str, string(bres))
			continue
		}
		t.Logf("Ok")
	}
}

func TestFlatFlagListEncoding(t *testing.T) {
	type Case struct {
		name string
		json string
		err  bool
	}
	cases := []Case{
		{
			name: "Empty",
			json: "[]",
		},
		{
			name: "Single Simple",
			json: `["-foo"]`,
		},
		{
			name: "Multiple Bad",
			json: `["-foo","+bar","+baz","-bash"]`,
		},
		{
			name: "Bad state (?)",
			json: `["?foo"]`,
			err:  true,
		},
		{
			name: "Bad state (~)",
			json: `["~foo"]`,
			err:  true,
		},
	}
	for _, c := range cases {
		str := fmt.Sprintf(`{"Flags":%s}`, c.json)
		var res struct {
			Flags spdl.FlatFlagList
		}
		err := json.Unmarshal([]byte(str), &res)
		if c.err {
			if err == nil {
				t.Fatalf("Expected a error, got nil")
			} else {
				t.Logf("Ok")
			}
			continue
		} else {
			if err != nil {
				t.Fatalf("Unexpected error %s", err.Error())
				continue
			}
		}

		bres, err := json.Marshal(&res)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
			continue
		}
		if str != string(bres) {
			t.Fatalf("Expected %s, got %s", str, string(bres))
			continue
		}
		t.Logf("Ok")
	}
}
