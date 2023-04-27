package asm

import (
    `sync/atomic`

    `github.com/chenzhuoyu/iasm/internal/tag`
)

// MaxOperands is the maximum number of operands an instruction can take.
const MaxOperands = 6

// Operands represents a sequence of operand required by an instruction.
type Operands [MaxOperands]interface{}

// BranchType represents the branch type of an instruction.
type BranchType uint8

const (
    BranchNone BranchType = iota
    BranchAlways
    BranchConditional
)

// InstructionDomain represents the domain of an instruction.
type InstructionDomain uint8

const (
    DomainGeneric InstructionDomain = iota
    DomainPseudo
    DomainArchSpecific
)

// Instruction represents an abstract unencoded instruction.
type Instruction struct {
    refs   int64
    arch   *Arch
    PC     uintptr
    Name   string
    Argc   int
    Argv   Operands
    Pseudo Pseudo
    Branch BranchType
    Domain InstructionDomain
}

func (self *Instruction) Free() {
    if atomic.AddInt64(&self.refs, -1) == 0 {
        self.clearArgs()
        self.clearPseudo()
        self.arch.impl.Free(self)
    }
}

func (self *Instruction) Encode() (m []byte) {
    self.arch.impl.Encode(self, &m)
    return
}

func (self *Instruction) Retain() *Instruction {
    atomic.AddInt64(&self.refs, 1)
    return self
}

func (self *Instruction) clearArgs() {
    for i := 0; i < self.Argc; i++ {
        if v, ok := self.Argv[i].(tag.Disposable); ok {
            v.Free()
        }
    }
}

func (self *Instruction) clearPseudo() {
    if self.Pseudo != nil {
        self.Pseudo.Free()
    }
}
