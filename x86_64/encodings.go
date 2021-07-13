package x86_64

import (
    `encoding/binary`
    `math`
)

/** Operand Encoding Helpers **/

func addr(v interface{}) interface{} {
    return v.(MemoryOperand).Addr
}

func offs(v interface{}) int64 {
    return int64(v.(RelativeOffset))
}

func lcode(v interface{}) byte {
    switch r := v.(type) {
        case Register8      : return byte(r & 0x07)
        case Register16     : return byte(r & 0x07)
        case Register32     : return byte(r & 0x07)
        case Register64     : return byte(r & 0x07)
        case KRegister      : return byte(r & 0x07)
        case MMRegister     : return byte(r & 0x07)
        case XMMRegister    : return byte(r & 0x07)
        case YMMRegister    : return byte(r & 0x07)
        case ZMMRegister    : return byte(r & 0x07)
        case MaskedRegister : return lcode(r.Reg)
        default             : panic("v is not a register")
    }
}

func hcode(v interface{}) byte {
    switch r := v.(type) {
        case Register8      : return byte(r >> 3) & 1
        case Register16     : return byte(r >> 3) & 1
        case Register32     : return byte(r >> 3) & 1
        case Register64     : return byte(r >> 3) & 1
        case KRegister      : return byte(r >> 3) & 1
        case MMRegister     : return byte(r >> 3) & 1
        case XMMRegister    : return byte(r >> 3) & 1
        case YMMRegister    : return byte(r >> 3) & 1
        case ZMMRegister    : return byte(r >> 3) & 1
        case MaskedRegister : return hcode(r.Reg)
        default             : panic("v is not a register")
    }
}

func toImmAny(v interface{}) int64 {
    if x, ok := asInt64(v); ok {
        return x
    } else {
        panic("value is not an integer")
    }
}

func toHcodeOpt(v interface{}) byte {
    if v == nil {
        return 0
    } else {
        return hcode(v)
    }
}

/** Instruction Encoding Helpers **/

const (
    _ACC0 = 0x01
    _ACC1 = 0x02
    _REL1 = 0x04
    _REL4 = 0x08
    _MRSD = 0x10
    _OREX = 0x20
    _VEX2 = 0x40
    _EVEX = 0x80
)

const (
    _MAX_INST = 16
)

type _Encoding struct {
    len     int
    flags   int
    bytes   [_MAX_INST]byte
    encoder func(m *_Encoding, v []interface{})
}

// buf ensures len + n <= len(bytes).
func (self *_Encoding) buf(n int) []byte {
    if i := self.len; i + n > _MAX_INST {
        panic("instruction too long")
    } else {
        return self.bytes[i:]
    }
}

// emit emits a single byte.
func (self *_Encoding) emit(v byte) {
    self.buf(1)[0] = v
    self.len++
}

// imm1 encodes a single byte immediate value.
func (self *_Encoding) imm1(v int64) {
    self.emit(byte(v))
}

// imm2 encodes a two-byte immediate value in little-endian.
func (self *_Encoding) imm2(v int64) {
    binary.LittleEndian.PutUint16(self.buf(2), uint16(v))
    self.len += 2
}

// imm4 encodes a 4-byte immediate value in little-endian.
func (self *_Encoding) imm4(v int64) {
    binary.LittleEndian.PutUint32(self.buf(4), uint32(v))
    self.len += 4
}

// imm8 encodes an 8-byte immediate value in little-endian.
func (self *_Encoding) imm8(v int64) {
    binary.LittleEndian.PutUint64(self.buf(8), uint64(v))
    self.len += 8
}

// vex2 encodes a 2-byte or 3-byte VEX prefix.
//
//                          2-byte VEX prefix:
// Requires: VEX.W = 0, VEX.mmmmm = 0b00001 and VEX.B = VEX.X = 0
//         +----------------+
// Byte 0: | Bits 0-7: 0xC5 |
//         +----------------+
//
//         +-----------+----------------+----------+--------------+
// Byte 1: | Bit 7: ~R | Bits 3-6 ~vvvv | Bit 2: L | Bits 0-1: pp |
//         +-----------+----------------+----------+--------------+
//
//                          3-byte VEX prefix:
//         +----------------+
// Byte 0: | Bits 0-7: 0xC4 |
//         +----------------+
//
//         +-----------+-----------+-----------+-------------------+
// Byte 1: | Bit 7: ~R | Bit 6: ~X | Bit 5: ~B | Bits 0-4: 0b00001 |
//         +-----------+-----------+-----------+-------------------+
//
//         +----------+-----------------+----------+--------------+
// Byte 2: | Bit 7: 0 | Bits 3-6: ~vvvv | Bit 2: L | Bits 0-1: pp |
//         +----------+-----------------+----------+--------------+
//
func (self *_Encoding) vex2(lpp byte, r byte, rm interface{}, vvvv byte) {
    var b byte
    var x byte

    /* VEX.R must be a single-bit mask */
    if r != 0 && r != 1 {
        panic("VEX.R must be a 1-bit mask")
    }

    /* VEX.Lpp must be a 3-bit mask */
    if lpp &^ 0b111 != 0 {
        panic("VEX.Lpp must be a 3-bit mask")
    }

    /* VEX.vvvv must be a 4-bit mask */
    if vvvv &^ 0b111 != 0 {
        panic("VEX.vvvv must be a 4-bit mask")
    }

    /* encode the RM byte if any */
    if rm != nil {
        switch v := rm.(type) {
            case Register       : b = hcode(v)
            case MemoryAddress  : b, x = toHcodeOpt(v.Base), toHcodeOpt(v.Index)
            case RelativeOffset : break
            default             : panic("rm is expected to be a register or a memory address")
        }
    }

    /* if VEX.B and VEX.X are zeroes, 2-byte VEX prefix can be used */
    if x == 0 && b == 0 {
        self.emit(0xc5)
        self.emit(0xf8 ^ (r << 7) ^ (vvvv << 3) ^ lpp)
    } else {
        self.emit(0xc6)
        self.emit(0xe1 ^ (r << 7) ^ (x << 6) ^ (b << 5))
        self.emit(0x78 ^ (vvvv << 3) ^ lpp)
    }
}

// rexm encodes a mandatory REX prefix
func (self *_Encoding) rexm(w byte, r byte, rm interface{}) {
    var b byte
    var x byte

    /* REX.R must be 0 or 1 */
    if r != 0 && r != 1 {
        panic("REX.R must be 0 or 1")
    }

    /* REX.W must be 0 or 1 */
    if w != 0 && w != 1 {
        panic("REX.W must be 0 or 1")
    }

    /* encode the RM byte */
    switch v := rm.(type) {
        case MemoryAddress  : b, x = toHcodeOpt(v.Base), toHcodeOpt(v.Index)
        case RelativeOffset : break
        default             : panic("rm is expected to be a register or a memory address")
    }

    /* encode the REX prefix */
    self.emit(0x40 | (w << 3) | (r << 2) | (x << 1) | b)
}

// rexo encodes an optional REX prefix.
func (self *_Encoding) rexo(r byte, rm interface{}, force bool) {
    var b byte
    var x byte

    /* REX.R must be 0 or 1 */
    if r != 0 && r != 1 {
        panic("REX.R must be 0 or 1")
    }

    /* encode the RM byte */
    switch v := rm.(type) {
        case Register       : b = hcode(v)
        case MemoryAddress  : b, x = toHcodeOpt(v.Base), toHcodeOpt(v.Index)
        case RelativeOffset : break
        default             : panic("rm is expected to be a register or a memory address")
    }

    /* if REX.R, REX.X, and REX.B are all zeroes, REX prefix can be omitted */
    if force || r != 0 || x != 0 || b != 0 {
        self.emit(0x40 | (r << 2) | (x << 1) | b)
    }
}

// mrsd encodes ModR/M, SIB, Displacement.
//
//                    ModR/M byte
// +----------------+---------------+---------------+
// | Bits 6-7: Mode | Bits 3-5: Reg | Bits 0-2: R/M |
// +----------------+---------------+---------------+
//
//                         SIB byte
// +-----------------+-----------------+----------------+
// | Bits 6-7: Scale | Bits 3-5: Index | Bits 0-2: Base |
// +-----------------+-----------------+----------------+
//
func (self *_Encoding) mrsd(reg byte, rm interface{}, disp8v int32) {
    var ok bool
    var mm MemoryAddress
    var ro RelativeOffset

    /* ModRM encodes the lower 3-bit of the register */
    if reg > 7 {
        panic("invalid register bits")
    }

    /* check the displacement scale */
    switch disp8v {
        case  1: break
        case  2: break
        case  4: break
        case  8: break
        case 16: break
        case 32: break
        case 64: break
        default: panic("invalid displacement size")
    }

    /* special case: RIP-relative offset
     * ModRM.Mode == 0 and ModeRM.R/M == 5 indicates (rip + disp32) addressing */
    if ro, ok = rm.(RelativeOffset); ok {
        self.emit(0x05 | (reg << 3))
        self.imm4(int64(ro))
        return
    }

    /* must be a generic memory address */
    if mm, ok = rm.(MemoryAddress); !ok {
        panic("rm must be a memory address")
    }

    /* global addressing is not yet supported */
    if mm.Base == nil && mm.Index == nil {
        panic("global addressing is not yet supported")
    }

    /* no SIB byte */
    if mm.Index == nil && lcode(mm.Base) != 0b100 {
        cc := lcode(mm.Base)
        dv := mm.Displacement

        /* ModRM.Mode == 0 (no displacement) */
        if dv == 0 && mm.Base != RBP && mm.Base != R13 {
            if cc == 0b101 {
                panic("rbp/r13 is not encodable as a base register (interpreted as disp32 address)")
            } else {
                self.emit((reg << 3) | cc)
                return
            }
        }

        /* ModRM.Mode == 1 (8-bit displacement) */
        if dq := dv / disp8v; dq >= math.MinInt8 && dq <= math.MaxInt8 && dv % disp8v == 0 {
            self.emit(0x40 | (reg << 3) | cc)
            self.imm1(int64(dq))
            return
        }

        /* ModRM.Mode == 2 (32-bit displacement) */
        self.emit(0x80 | (reg << 3) | cc)
        self.imm4(int64(mm.Displacement))
        return
    }

    /* all encodings below use ModRM.R/M = 4 (0b100) to indicate the presence of SIB */
    if mm.Index == RSP {
        panic("rsp is not encodable as an index register (interpreted as no index)")
    }

    /* index = 4 (0b100) denotes no-index encoding */
    var scale byte
    var index byte = 0x04

    /* encode the scale byte */
    if mm.Scale != 0 {
        switch mm.Scale {
            case 1  : scale = 0
            case 2  : scale = 1
            case 4  : scale = 2
            case 8  : scale = 3
            default : panic("invalid scale value")
        }
    }

    /* encode the index byte */
    if mm.Index != nil {
        index = lcode(mm.Index)
    }

    /* SIB.Base = 5 (0b101) and ModRM.Mode = 0 indicates no-base encoding with disp32 */
    if mm.Base == nil {
        self.emit((reg << 3) | 0b100)
        self.emit((scale << 6) | (index << 3) | 0b101)
        self.imm4(int64(mm.Displacement))
        return
    }

    /* base L-code & displacement value */
    cc := lcode(mm.Base)
    dv := mm.Displacement

    /* ModRM.Mode == 0 (no displacement) */
    if dv == 0 && cc != 0b101 {
        self.emit((reg << 3) | 0b100)
        self.emit((scale << 6) | (index << 3) | cc)
        return
    }

    /* ModRM.Mode == 1 (8-bit displacement) */
    if dq := dv / disp8v; dq >= math.MinInt8 && dq <= math.MaxInt8 && dv % disp8v == 0 {
        self.emit(0x44 | (reg << 3))
        self.emit((scale << 6) | (index << 3) | cc)
        self.imm1(int64(dq))
        return
    }

    /* ModRM.Mode == 2 (32-bit displacement) */
    self.emit(0x84 | (reg << 3))
    self.emit((scale << 6) | (index << 3) | cc)
    self.imm4(int64(mm.Displacement))
}

// encode invokes the encoder to encode this instruction.
func (self *_Encoding) encode(v []interface{}) int {
    self.len = 0
    self.encoder(self, v)
    return self.len
}
