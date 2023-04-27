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
    pc := self.PC
    iv := uint32(0)

    /* check for pseudo-instructions */
    if self.Pseudo != nil {
        return self.Pseudo.Encode(m, pc)
    }

    /* check for dry run */
    if m == nil {
        return _INS_LEN
    }

    /* encode the instruction */
    if self.instr != 0 {
        iv = self.instr
    } else if self.encoder != nil {
        iv = self.encoder(pc)
    } else{
        panic("aarch64: uninitialized instruction")
    }

    /* add to buffer */
    *m = binary.LittleEndian.AppendUint32(*m, iv)
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

// LDI is a pseudo-instruction that loads a 64-bit immediate into a register.
//
// The LDI pseudo-instruction generates the most efficient single instruction
// for a specific constant:
//
//   * If the constant can be constructed with a single MOV or MVN instruction,
//     the assembler generates the appropriate instruction.
//
//   * If not, the assembler will encode the immediate with MOVZ / MOVK pairs.
//
// NOTE: This pseudo-instruction may expand to multiple instructions, so it
//       does not return the instruction instance.
//
func (self *Program) LDI(reg asm.Register, val uint64) {
    var z int
    var v [4]uint16

    /* simple case if the immediate is encodable with "MOV Rd, #imm" */
    if isXr(reg) && (isMOVxImm(val, 64, true) || isMOVxImm(val, 64, false)) ||
       isWr(reg) && (isMOVxImm(val, 32, true) || isMOVxImm(val, 32, false)) ||
       isXrOrSP(reg) && isMask64(val) ||
       isWrOrWSP(reg) && isMask32(val) {
        self.MOV(reg, val)
        return
    }

    /* decompose the immediate */
    v[0] = uint16(val)
    v[1] = uint16(val >> 16)
    v[2] = uint16(val >> 32)
    v[3] = uint16(val >> 48)

    /* encode with MOVZ/MOVK pairs */
    for i, x := range v {
        if x != 0 {
            if z++; z == 1 {
                self.MOVZ(reg, x, LSL(i * 16))
            } else {
                self.MOVK(reg, x, LSL(i * 16))
            }
        }
    }
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
