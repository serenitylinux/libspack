package spdl

import (
	"reflect"
	"testing"
)

func TestDepParse(t *testing.T) {
	type Case struct {
		name   string
		input  string
		expect Dep
		err    bool
	}

	var and op = And
	var or op = Or

	cases := []Case{
		{
			name:  "Basic",
			input: "basic",
			expect: Dep{
				Name: "basic",
			},
		},
		{
			name:  "Basic Condition",
			input: "[-cond] basic",
			expect: Dep{
				Condition: &ExprList{e: expr{flag: FlatFlag{Name: "cond", Enabled: false}}},
				Name:      "basic",
			},
		},
		{
			name:  "Advanced Condition",
			input: "[-cond && [-baz || +bar]] advanced",
			expect: Dep{
				Condition: &ExprList{
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
				Name: "advanced",
			},
		},
		{
			name:  "Bad Condition",
			input: "[?cond] basic",
			err:   true,
		},
		{
			name:  "Version Basic",
			input: "basic>=3",
			expect: Dep{
				Name: "basic",
				Version1: &Version{
					typ: GT,
					ver: "3",
				},
			},
		},
		{
			name:  "Version Advanced",
			input: "basic >=3ac <=10.4b",
			expect: Dep{
				Name: "basic",
				Version1: &Version{
					typ: GT,
					ver: "3ac",
				},
				Version2: &Version{
					typ: LT,
					ver: "10.4b",
				},
			},
		},
		{
			name:  "Basic Flags",
			input: "basic(+dev -doc ?foo ~bar)",
			expect: Dep{
				Name: "basic",
				Flags: &FlagList{
					"dev": Flag{Name: "dev", State: Enabled},
					"doc": Flag{Name: "doc", State: Disabled},
					"foo": Flag{Name: "foo", State: Inherit},
					"bar": Flag{Name: "bar", State: Invert},
				},
			},
		},
		{
			name:  "Somewhat Advanced Flags",
			input: "basic(+dev ~doc(+foo))",
			expect: Dep{
				Name: "basic",
				Flags: &FlagList{
					"dev": Flag{Name: "dev", State: Enabled},
					"doc": Flag{
						Name:  "doc",
						State: Invert,
						Expr: &ExprList{
							e: expr{flag: FlatFlag{Name: "foo", Enabled: true}},
						},
					},
				},
			},
		},
		{
			name:  "Advanced Flags",
			input: "basic(+dev ~doc(+foo || [-baza && +build]) ?other(-bazh))",
			expect: Dep{
				Name: "basic",
				Flags: &FlagList{
					"dev": Flag{Name: "dev", State: Enabled},
					"doc": Flag{
						Name:  "doc",
						State: Invert,
						Expr: &ExprList{
							e:  expr{flag: FlatFlag{Name: "foo", Enabled: true}},
							op: &or,
							next: &ExprList{
								e: expr{
									list: &ExprList{
										e:  expr{flag: FlatFlag{Name: "baza", Enabled: false}},
										op: &and,
										next: &ExprList{
											e: expr{flag: FlatFlag{Name: "build", Enabled: true}},
										},
									},
								},
							},
						},
					},
					"other": Flag{
						Name:  "other",
						State: Inherit,
						Expr: &ExprList{
							e: expr{flag: FlatFlag{Name: "bazh", Enabled: false}},
						},
					},
				},
			},
		},
		{
			name:  "ALL THE THINGS!",
			input: "[-cond && [-baz || +bar]] basic >=3ac <=10.4b (+dev ~doc(+foo || [-baza && +build]) ?other(-bazh))",
			expect: Dep{
				Condition: &ExprList{
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
				Name: "basic",
				Version1: &Version{
					typ: GT,
					ver: "3ac",
				},
				Version2: &Version{
					typ: LT,
					ver: "10.4b",
				},
				Flags: &FlagList{
					"dev": Flag{Name: "dev", State: Enabled},
					"doc": Flag{
						Name:  "doc",
						State: Invert,
						Expr: &ExprList{
							e:  expr{flag: FlatFlag{Name: "foo", Enabled: true}},
							op: &or,
							next: &ExprList{
								e: expr{
									list: &ExprList{
										e:  expr{flag: FlatFlag{Name: "baza", Enabled: false}},
										op: &and,
										next: &ExprList{
											e: expr{flag: FlatFlag{Name: "build", Enabled: true}},
										},
									},
								},
							},
						},
					},
					"other": Flag{
						Name:  "other",
						State: Inherit,
						Expr: &ExprList{
							e: expr{flag: FlatFlag{Name: "bazh", Enabled: false}},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Log(c.name)
		actual, err := ParseDep(c.input)
		if c.err {
			if err == nil {
				t.Errorf("Expected error")
			} else {
				t.Logf("Ok")
			}
			continue
		} else if err != nil {
			t.Errorf("Unexpected error %v", err.Error())
			continue
		}

		if !reflect.DeepEqual(actual, c.expect) {
			t.Errorf("Expected \n%+v\ngot \n%+v", c.expect, actual)
			continue
		}
		t.Logf("Ok")
	}
}
