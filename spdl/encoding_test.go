package spdl

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
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
			Flags FlagList
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
		t.Log(c.name)
		str := fmt.Sprintf(`{"Flags":%s}`, c.json)
		var res struct {
			Flags FlatFlagList
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

func TestDepEncoding(t *testing.T) {
	type Case struct {
		name   string
		json   string
		expect Dep
	}
	cases := []Case{
		{
			name:   "Basic",
			json:   "foo",
			expect: Dep{Name: "foo"},
		},
		{
			name:   "ALL THE THINGS!",
			json:   fulldepJSON,
			expect: fulldep,
		},
	}
	for _, c := range cases {
		t.Log(c.name)
		str := fmt.Sprintf(`{"Dep":"%s"}`, c.json)
		var res struct {
			Dep Dep
		}
		err := json.Unmarshal([]byte(str), &res)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			continue
		}

		if !reflect.DeepEqual(c.expect, res.Dep) {
			t.Errorf("Expected \n%+v\n got \n%+v", c.expect, res.Dep)
		}

		//Re-Encode
		bres, err := json.Marshal(&res)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			continue
		}

		res.Dep = Dep{}
		err = json.Unmarshal(bres, &res)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			continue
		}

		if !reflect.DeepEqual(c.expect, res.Dep) {
			t.Errorf("Expected \n%+v\n got \n%+v", c.expect, res.Dep)
		}

		t.Logf("Ok")
	}
}

func TestFlagExprEncoding(t *testing.T) {
	type Case struct {
		name   string
		json   string
		expect FlagExpr
	}

	cases := []Case{
		{
			name: "Basic",
			json: "+foo",
			expect: FlagExpr{
				Flag: FlatFlag{Name: "foo", Enabled: true},
			},
		},
		{
			name: "Advanced",
			json: "+foo(-cond && [-baz || +bar])",
			expect: FlagExpr{
				Flag: FlatFlag{Name: "foo", Enabled: true},
				list: &ExprList{
					e:  expr{flag: FlatFlag{Name: "cond", Enabled: false}},
					op: &and,
					next: &ExprList{
						e: expr{
							list: &ExprList{
								e:  expr{flag: FlatFlag{Name: "baz", Enabled: false}},
								op: &or,
								next: &ExprList{
									e: expr{flag: FlatFlag{Name: "bar", Enabled: true}},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Log(c.name)

		str := fmt.Sprintf(`{"FlagExpr":"%s"}`, c.json)
		var res struct {
			FlagExpr FlagExpr
		}
		err := json.Unmarshal([]byte(str), &res)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			continue
		}

		if !reflect.DeepEqual(c.expect, res.FlagExpr) {
			t.Errorf("Expected1 \n%+v\n got \n%+v", c.expect, res.FlagExpr)
		}

		//Re-Encode
		bres, err := json.Marshal(&res)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			continue
		}

		res.FlagExpr = FlagExpr{}
		err = json.Unmarshal(bres, &res)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			continue
		}

		if !reflect.DeepEqual(c.expect, res.FlagExpr) {
			t.Errorf("Expected \n%+v\n got \n%+v", c.expect, res.FlagExpr)
		}
		t.Log("Ok")
	}
}
