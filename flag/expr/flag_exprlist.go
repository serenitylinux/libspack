package expr

import (
	"github.com/serenitylinux/libspack/flag"
	"github.com/serenitylinux/libspack/parser"
)

type ExprList struct {
	e    expr
	op   *op
	next *ExprList
}

func ParseExprList(in *parser.Input) (*ExprList, error) {
	list := new(ExprList)

	e, err := parseExpr(in)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, nil
	}
	list.e = *e

	if op_isnext(in) {
		nop, err := parseOp(in)
		if err != nil {
			return nil, err
		}

		nel, err := ParseExprList(in)
		if err != nil {
			return nil, err
		}

		list.op = nop
		list.next = nel
	}
	return list, nil
}
func (list *ExprList) Enabled(flist flag.FlatFlagList) bool {
	if list == nil {
		return true
	}
	if list.op == nil {
		return list.e.verify(flist)
	}
	if *list.op == And {
		return list.e.verify(flist) && list.next.Enabled(flist)
	} else {
		return list.e.verify(flist) || list.next.Enabled(flist)
	}
}
func (list *ExprList) String() string {
	if list == nil {
		return ""
	}
	if list.op == nil {
		return list.e.String()
	}
	return list.e.String() + list.op.String() + list.next.String()
}
