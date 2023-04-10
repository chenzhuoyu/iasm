package aarch64

import (
    `encoding/binary`

    `github.com/chenzhuoyu/iasm/asm`
)

const (
    _INS_LEN = 4
)

type Instruction struct {
    asm.Instruction
    instr   uint32
    encoder func(uintptr) uint32
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

func (self *Instruction) append(m *[]byte) int {
    *m = binary.LittleEndian.AppendUint32(*m, self.encode(self.PC))
    return _INS_LEN
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

type Program struct {
    asm.Program
}

func (self *Program) alloc(name string, argc int, argv asm.Operands) *Instruction {
    return this(self.Append(name, argc, argv))
}

func (self *Program) assemble(pc uintptr) []byte {
    // TODO implement me
    panic("implement me")
}
