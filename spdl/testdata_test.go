package spdl

var and op = And
var or op = Or

func buildFlagListPtr(fs ...Flag) *FlagList {
	res := buildFlagList(fs...)
	return &res
}

func buildFlagList(fs ...Flag) FlagList {
	fl := NewFlagList(len(fs))
	for _, f := range fs {
		fl.Add(f)
	}
	return fl
}

func buildFlatFlagList(fs ...FlatFlag) FlatFlagList {
	fl := NewFlatFlagList(len(fs))
	for _, f := range fs {
		fl.Add(f)
	}
	return fl
}

var fulldepJSON = "[-cond && (-baz || +bar)] basic >=3ac <=10.4b (+dev ~doc(+foo || (-baza && +build)) ?other(-bazh))"
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
	Flags: buildFlagListPtr(
		Flag{Name: "dev", State: Enabled},
		Flag{
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
		Flag{
			Name:  "other",
			State: Inherit,
			Expr: &ExprList{
				e: expr{flag: FlatFlag{Name: "bazh", Enabled: false}},
			},
		},
	),
}
