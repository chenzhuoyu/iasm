package asm

import (
    `github.com/chenzhuoyu/iasm/expr`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Instruction defines an abstract instruction.
type Instruction interface {
    tag.Disposable
    PC() uintptr
    Retain() Instruction
    Mnemonic() string
    Operands() []interface{}
}

// Program is a sequence of instructions.
type Program interface {
    // Disposable makes the Program disposable.
    //
    // NOTE: This also frees all the instructions, labels, memory
    //       operands and expressions associated with this program.
    //
    tag.Disposable

    // Link pins a label at the current position.
    Link(p *Label)

    // Data is a pseudo-instruction that adds raw bytes to the assembled code.
    Data(v []byte) Instruction

    // Byte is a pseudo-instruction that adds raw byte to the assembled code.
    Byte(v *expr.Expr) Instruction

    // Word is a pseudo-instruction that adds raw uint16 as little-endian to the assembled code.
    Word(v *expr.Expr) Instruction

    // Long is a pseudo-instruction that adds raw uint32 as little-endian to the assembled code.
    Long(v *expr.Expr) Instruction

    // Quad is a pseudo-instruction that adds raw uint64 as little-endian to the assembled code.
    Quad(v *expr.Expr) Instruction

    // Align is a pseudo-instruction that ensures the PC is aligned to a certain value.
    Align(align uint64, padding *expr.Expr) Instruction

    // Assemble assembles and links the entire program into machine code.
    Assemble(pc uintptr) []byte
}

// AssembleAndFree is like Program.Assemble, but it frees the Program after assembling.
func AssembleAndFree(p Program, pc uintptr) (r []byte) {
    r = p.Assemble(pc)
    p.Free()
    return
}
