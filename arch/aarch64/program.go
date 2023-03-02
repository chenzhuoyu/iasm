package aarch64

import (
    `sync/atomic`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Operands represents a sequence of operand required by an instruction.
type Operands [_N_args]interface{}

// Instruction represents an unencoded instruction.
type Instruction struct {
    next    *Instruction
    refs    int64
    pc      uintptr
    name    string
    argc    int
    argv    Operands
    instr   uint32
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

func (self *Instruction) Name() string {
    return self.name
}

func (self *Instruction) Retain() asm.Instruction {
    atomic.AddInt64(&self.refs, 1)
    return self
}

func (self *Instruction) Operands() []interface{} {
    return self.argv[:self.argc]
}

// Program represents a sequence of instructions.
type Program struct {
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

func (self *Program) Free() {
    self.clear()
    freeProgram(self)
}

func (self *Program) Link(p *asm.Label) {
    // TODO implement me
    panic("implement me")
}

func (self *Program) Assemble(pc uintptr) []byte {
    // TODO implement me
    panic("implement me")
}
