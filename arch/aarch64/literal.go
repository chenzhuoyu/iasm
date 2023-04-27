package aarch64

import (
    `fmt`
)

type (
    _LitInt   uint64
	_LitName  string
    _LitFloat float64
)

const (
    _LitSym = "$_LitSym"
)

func (self _LitInt)   String() string { return fmt.Sprintf("LitInt(%d)", self) }
func (self _LitName)  String() string { return fmt.Sprintf("LitName(%s)", string(self)) }
func (self _LitFloat) String() string { return fmt.Sprintf("LitFloat(%f)", self) }
