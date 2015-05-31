package spdl

//TODO doc on Serenity Package Dependency Language (SPDL)

/*
FlagExpr
Represents a flag state and a possible expression dependency
-dev([+qt && -gtk] || [-qt && +gtk])
[is_enabled_default]name(deps)

exprlist  = expr + exprlist'
exprlist' = arg + exprlist || \0

expr = sub || flag
arg = '&&,||'

sub = '[' + exprlist + ']'
*/
