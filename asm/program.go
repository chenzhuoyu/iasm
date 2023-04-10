package asm

import (
    `github.com/chenzhuoyu/iasm/expr`
)

// Program represents a sequence of instructions.
type Program struct {
    Arch  *Arch
    Instr []*Instruction
}

func (self *Program) clear() {
    for _, p := range self.Instr {
        p.Free()
    }
}

// Free returns the Program object into pool.
// Any operation performed after Free is undefined behavior.
//
// NOTE: This also frees all the instructions, labels, memory
//       operands and expressions associated with this program.
//
func (self *Program) Free() {
    self.clear()
    freeProgram(self)
}

// Link pins a label at the current position.
func (self *Program) Link(p *Label) {
    if p.Dest != nil {
        panic("lable was alreay linked")
    } else {
        p.Dest = self.Pseudo(Nop).Retain()
    }
}

// Data is a pseudo-instruction that adds raw bytes to the assembled code.
func (self *Program) Data(v []byte) *Instruction {
    p := self.Pseudo(Data)
    p.Pseudo.Data = v
    return p
}

// Byte is a pseudo-instruction that adds raw byte to the assembled code.
func (self *Program) Byte(v *expr.Expr) *Instruction {
    p := self.Pseudo(Byte)
    p.Pseudo.Expr = v
    return p
}

// Word is a pseudo-instruction that adds raw uint16 as little-endian to the assembled code.
func (self *Program) Word(v *expr.Expr) *Instruction {
    p := self.Pseudo(Word)
    p.Pseudo.Expr = v
    return p
}

// Long is a pseudo-instruction that adds raw uint32 as little-endian to the assembled code.
func (self *Program) Long(v *expr.Expr) *Instruction {
    p := self.Pseudo(Long)
    p.Pseudo.Expr = v
    return p
}

// Quad is a pseudo-instruction that adds raw uint64 as little-endian to the assembled code.
func (self *Program) Quad(v *expr.Expr) *Instruction {
    p := self.Pseudo(Quad)
    p.Pseudo.Expr = v
    return p
}

// Align is a pseudo-instruction that ensures the PC is aligned to a certain value.
func (self *Program) Align(to uint64, padding *expr.Expr) *Instruction {
    p := self.Pseudo(Align)
    p.Pseudo.Uint = to
    p.Pseudo.Expr = padding
    return p
}

// Pseudo appends a pesudo-instruction with specific type to the instruction buffer.
func (self *Program) Pseudo(kind PseudoType) (p *Instruction) {
    p = self.Append(kind.String(), 0, Operands{})
    p.Domain = DomainPseudo
    p.Pseudo.Kind = kind
    return
}

// Append allocates a new instruction with name and argv:argc, and adds to the instruction buffer.
func (self *Program) Append(name string, argc int, argv Operands) *Instruction {
    p := self.Arch.impl.New()
    p.refs = 1
    p.arch = self.Arch
    p.Name = name
    p.Argc = argc
    p.Argv = argv
    self.Instr = append(self.Instr, p)
    return p
}

// Assemble assembles and links the entire program into machine code.
func (self *Program) Assemble(pc uintptr) (ret []byte) {
    return self.Arch.impl.Assemble(self, pc)
}

// AssembleAndFree is like Assemble, but it frees the Program after assembling.
func (self *Program) AssembleAndFree(pc uintptr) (ret []byte) {
    ret = self.Assemble(pc)
    self.Free()
    return
}