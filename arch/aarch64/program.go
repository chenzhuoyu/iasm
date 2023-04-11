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

func (self *Instruction) encode(m *[]byte) int {
    inst := uint32(0)
    kind := self.Pseudo.Kind

    /* check for pseudo-instructions */
    if kind != 0 {
        return self.Pseudo.Encode(m, self.PC)
    }

    /* check for dry run */
    if m == nil {
        return _INS_LEN
    }

    /* encode the instruction */
    if self.instr != 0 {
        inst = self.instr
    } else if self.encoder != nil {
        inst = self.encoder(self.PC)
    } else{
        panic("aarch64: uninitialized instruction")
    }

    /* add to buffer */
    *m = binary.LittleEndian.AppendUint32(*m, inst)
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

func (self *Program) encode(m *[]byte) {
    for _, p := range self.Instr {
        this(p).encode(m)
    }
}

func (self *Program) assemble(pc uintptr) []byte {
    orig := pc
    inst := self.Instr

    /* pre-compute PC for every instruction */
    for _, p := range inst {
        p.PC = pc
        pc += uintptr(this(p).encode(nil))
    }

    /* allocate space for machine code */
    nb := pc - orig
    ret := make([]byte, 0, nb)

    /* encode every instruction */
    self.encode(&ret)
    return ret
}
