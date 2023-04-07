package asm

import (
    `fmt`
    `math`
    `math/bits`

    `github.com/chenzhuoyu/iasm/expr`
)

// PseudoType is the type of pseudo-instructions.
type PseudoType int

const (
    Invalid PseudoType = iota
    Nop
    Data
    Byte
    Word
    Long
    Quad
    Align
)

func (self PseudoType) String() string {
    switch self {
        case Nop   : return ".nop"
        case Data  : return ".data"
        case Byte  : return ".byte"
        case Word  : return ".word"
        case Long  : return ".long"
        case Quad  : return ".quad"
        case Align : return ".align"
        default    : panic("unreachable")
    }
}

// Pseudo represents one of the pseudo-instructions.
type Pseudo struct {
    Kind PseudoType
    Data []byte
    Uint uint64
    Expr *expr.Expr
}

// Free destroyes the pseudo-instruction.
func (self *Pseudo) Free() {
    if self.Expr != nil {
        self.Expr.Free()
    }
}

// Encode attempts to encode the pseudo-instruction into m, at pc, and returns the encoded size.
func (self *Pseudo) Encode(m *[]byte, pc uintptr) int {
    switch self.Kind {
        case Nop   : return 0
        case Data  : self.encodeData(m)      ; return len(self.Data)
        case Byte  : self.encodeByte(m)      ; return 1
        case Word  : self.encodeWord(m)      ; return 2
        case Long  : self.encodeLong(m)      ; return 4
        case Quad  : self.encodeQuad(m)      ; return 8
        case Align : self.encodeAlign(m, pc) ; return self.alignSize(pc)
        default    : panic("invalid pseudo instruction")
    }
}

func (self *Pseudo) evalExpr(low int64, high int64) int64 {
    if v, err := self.Expr.Evaluate(); err != nil {
        panic(err)
    } else if v < low || v > high {
        panic(fmt.Sprintf("expression out of range [%d, %d]: %d", low, high, v))
    } else {
        return v
    }
}

func (self *Pseudo) alignSize(pc uintptr) int {
    if !ispow2(self.Uint) {
        panic(fmt.Sprintf("aligment should be a power of 2, not %d", self.Uint))
    } else {
        return align(int(pc), bits.TrailingZeros64(self.Uint)) - int(pc)
    }
}

func (self *Pseudo) encodeData(m *[]byte) {
    if m != nil {
        *m = append(*m, self.Data...)
    }
}

func (self *Pseudo) encodeByte(m *[]byte) {
    if m != nil {
        append8(m, byte(self.evalExpr(math.MinInt8, math.MaxUint8)))
    }
}

func (self *Pseudo) encodeWord(m *[]byte) {
    if m != nil {
        append16(m, uint16(self.evalExpr(math.MinInt16, math.MaxUint16)))
    }
}

func (self *Pseudo) encodeLong(m *[]byte) {
    if m != nil {
        append32(m, uint32(self.evalExpr(math.MinInt32, math.MaxUint32)))
    }
}

func (self *Pseudo) encodeQuad(m *[]byte) {
    if m != nil {
        if v, err := self.Expr.Evaluate(); err != nil {
            panic(err)
        } else {
            append64(m, uint64(v))
        }
    }
}

func (self *Pseudo) encodeAlign(m *[]byte, pc uintptr) {
    if m != nil {
        if self.Expr == nil {
            expandmm(m, self.alignSize(pc), 0)
        } else {
            expandmm(m, self.alignSize(pc), byte(self.evalExpr(math.MinInt8, math.MaxUint8)))
        }
    }
}
