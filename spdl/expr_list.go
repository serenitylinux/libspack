package spdl

import (
	"github.com/serenitylinux/libspack/parser"
)

// Linked list representing a flat flag expression

type exprlist struct {
	e    expr
	op   *op
	next *exprlist
}

func parseexprlist(in *parser.Input) (*exprlist, error) {
	list := new(exprlist)

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

		nel, err := parseexprlist(in)
		if err != nil {
			return nil, err
		}

		list.op = nop
		list.next = nel
	}
	return list, nil
}
func (list *exprlist) Enabled(flist FlatFlagList) bool {
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
func (list *exprlist) String() string {
	if list == nil {
		return ""
	}
	if list.op == nil {
		return list.e.String()
	}
	return list.e.String() + list.op.String() + list.next.String()
}
