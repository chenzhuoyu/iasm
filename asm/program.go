package asm

import (
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Instruction defines an abstract instruction.
type Instruction interface {
    tag.Disposable
    PC() uintptr
    Name() string
    Retain() Instruction
    Operands() []interface{}
}

// Program is a sequence of instructions.
type Program interface {
    tag.Disposable
    Link(p *Label)
    Assemble(pc uintptr) []byte
}

// AssembleAndFree is like Program.Assemble, but it frees the Program after assembling.
func AssembleAndFree(p Program, pc uintptr) (r []byte) {
    r = p.Assemble(pc)
    p.Free()
    return
}
