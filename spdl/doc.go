package spdl

//TODO doc on Serenity Package Dependency Language (SPDL)

/*
FlagExpr
Represents a flag state and a possible expression dependency
-dev([+qt && -gtk] || [-qt && +gtk])
[is_enabled_default]name(deps)

ExprList  = expr + ExprList'
ExprList' = arg + ExprList || \0

expr = sub || flag
arg = '&&,||'

sub = '[' + ExprList + ']'
*/
