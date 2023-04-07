package aarch64

import (
    `sync/atomic`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/expr`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Operands represents a sequence of operand required by an instruction.
type Operands [_N_args]interface{}

// InstructionClass represents one of the instruction classes.
type InstructionClass uint8

const (
    ClassGeneral InstructionClass = iota
    ClassFloat
    ClassFpSimd
    ClassAdvSimd
    ClassSVE
    ClassSVE2
    ClassSME
    ClassSME2
    ClassSystem
    ClassPseudo
)

// Instruction represents an unencoded instruction.
type Instruction struct {
    next    *Instruction
    refs    int64
    pc      uintptr
    name    string
    argc    int
    argv    Operands
    class   InstructionClass
    instr   uint32
    pseudo  asm.Pseudo
    encoder func(uintptr) uint32
}

func (self *Instruction) clear() {
    for i := 0; i < self.argc; i++ {
        if v, ok := self.argv[i].(tag.Disposable); ok {
            v.Free()
        }
    }
}

func (self *Instruction) encode(pc uintptr) uint32 {
    if self.instr != 0 {
        return self.instr
    } else if self.encoder != nil {
        return self.encoder(pc)
    } else{
        panic("aarch64: uninitialized instruction")
    }
}

func (self *Instruction) setins(ins uint32) *Instruction {
    self.instr = ins
    self.encoder = nil
    return self
}

func (self *Instruction) setenc(enc func(uintptr) uint32) *Instruction {
    self.instr = 0
    self.encoder = enc
    return self
}

func (self *Instruction) PC() uintptr {
    return self.pc
}

func (self *Instruction) Free() {
    if atomic.AddInt64(&self.refs, -1) == 0 {
        self.clear()
        freeInstruction(self)
    }
}

func (self *Instruction) Class() InstructionClass {
    return self.class
}

func (self *Instruction) Retain() asm.Instruction {
    atomic.AddInt64(&self.refs, 1)
    return self
}

func (self *Instruction) Mnemonic() string {
    return self.name
}

func (self *Instruction) Operands() []interface{} {
    return self.argv[:self.argc]
}

// Program represents a sequence of instructions.
type Program struct {
    arch *asm.Arch
    head *Instruction
    tail *Instruction
}

func (self *Program) clear() {
    for p, q := self.head, self.head; p != nil; p = q {
        q = p.next
        p.Free()
    }
}

func (self *Program) alloc(name string, argc int, argv Operands) *Instruction {
    p := self.tail
    q := newInstruction(name, argc, argv)

    /* attach to tail if any */
    if p != nil {
        p.next = q
    } else {
        self.head = q
    }

    /* set the new tail */
    self.tail = q
    return q
}

func (self *Program) pseudo(kind asm.PseudoType) (p *Instruction) {
    p = self.alloc(kind.String(), 0, Operands{})
    p.class = ClassPseudo
    p.pseudo.Kind = kind
    return
}

func (self *Program) require(ft Feature) {
    if !self.arch.HasFeature(ft) {
        panic("Feature '" + ft.String() + "' was not enabled")
    }
}

func (self *Program) Free() {
    self.clear()
    freeProgram(self)
}

func (self *Program) Link(p *asm.Label) {
    if p.Dest != nil {
        panic("lable was alreay linked")
    } else {
        p.Dest = self.pseudo(asm.Nop).Retain()
    }
}

func (self *Program) Data(v []byte) asm.Instruction {
    p := self.pseudo(asm.Data)
    p.pseudo.Data = v
    return p
}

func (self *Program) Byte(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Byte)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Word(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Word)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Long(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Long)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Quad(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Quad)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Align(align uint64, padding *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Align)
    p.pseudo.Uint = align
    p.pseudo.Expr = padding
    return p
}

func (self *Program) Assemble(pc uintptr) []byte {
    // TODO implement me
    panic("implement me")
}
