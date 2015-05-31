package spdl

var and op = And
var or op = Or

var fulldepJSON = "[-cond && [-baz || +bar]] basic >=3ac <=10.4b (+dev ~doc(+foo || [-baza && +build]) ?other(-bazh))"
var fulldep = Dep{
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
}
