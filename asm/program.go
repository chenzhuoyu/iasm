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
    return self.Pseudo(Data(v))
}

// Byte is a pseudo-instruction that adds raw byte to the assembled code.
func (self *Program) Byte(v *expr.Expr) *Instruction {
    return self.Pseudo(Byte(v))
}

// Word is a pseudo-instruction that adds raw uint16 as little-endian to the assembled code.
func (self *Program) Word(v *expr.Expr) *Instruction {
    return self.Pseudo(Word(v))
}

// Long is a pseudo-instruction that adds raw uint32 as little-endian to the assembled code.
func (self *Program) Long(v *expr.Expr) *Instruction {
    return self.Pseudo(Long(v))
}

// Quad is a pseudo-instruction that adds raw uint64 as little-endian to the assembled code.
func (self *Program) Quad(v *expr.Expr) *Instruction {
    return self.Pseudo(Quad(v))
}

// Align is a pseudo-instruction that ensures the PC is aligned to a certain value.
func (self *Program) Align(to uint64, padding *expr.Expr) *Instruction {
    return self.Pseudo(Align(to, padding))
}

// Pseudo appends a pesudo-instruction with specific type to the instruction buffer.
func (self *Program) Pseudo(ps Pseudo) (p *Instruction) {
    p = self.Append(ps.Type().String(), 0, Operands{})
    p.Pseudo = ps
    p.Domain = DomainPseudo
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