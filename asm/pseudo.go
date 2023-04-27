package asm

import (
    `fmt`
    `math`
    `math/bits`
    `strconv`
    `sync`
    `unsafe`

    `github.com/chenzhuoyu/iasm/expr`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// PseudoType is the type of pseudo-instructions.
type PseudoType int

const (
    PseudoNop PseudoType = iota
    PseudoData
    PseudoByte
    PseudoWord
    PseudoLong
    PseudoQuad
    PseudoAlign
)

// PseudoNames records the name for every pseudo-instruction type.
var PseudoNames = map[PseudoType]string {
    PseudoNop   : ".nop",
    PseudoData  : ".data",
    PseudoByte  : ".byte",
    PseudoWord  : ".word",
    PseudoLong  : ".long",
    PseudoQuad  : ".quad",
    PseudoAlign : ".align",
}

func (self PseudoType) String() string {
    if s, ok := PseudoNames[self]; ok {
        return s
    } else {
        return fmt.Sprintf("PseudoType(%d)", int(self))
    }
}

// Pseudo represents one of the pseudo-instructions.
type Pseudo interface {
    fmt.Stringer
    tag.Disposable
    Type() PseudoType
    Encode(m *[]byte, pc uintptr) int
}

type (
	_Nop struct{}
)

// Nop is the do-nothing instruction, it also encodes to nothing.
var Nop _Nop

func (_Nop) Free()                              {}
func (_Nop) Type()                   PseudoType { return PseudoNop }
func (_Nop) String()                 string     { return ".nop" }
func (_Nop) Encode(*[]byte, uintptr) int        { return 0 }

// Data represents a raw stream of bytes, encoded as is.
type Data []byte

func (self Data) Free() {}
func (self Data) Type() PseudoType { return PseudoData }

func (self Data) String() string {
    return fmt.Sprintf(`.data %s`, strconv.Quote(string(self)))
}

func (self Data) Encode(m *[]byte, _ uintptr) int {
    if m != nil { *m = append(*m, self...) }
    return len(self)
}

type _Literal struct {
    nb int
    lo int64
    hi int64
    ex *expr.Expr
    ty PseudoType
}

var (
    _literalPool sync.Pool
)

func newLit(ty PseudoType, lo int64, hi int64, nb int, val *expr.Expr) *_Literal {
    var p *_Literal
    var v interface{}

    /* attempt to grab one from pool */
    if v = _literalPool.Get(); v != nil {
        p = v.(*_Literal)
    } else {
        p = new(_Literal)
    }

    /* initialize the literal */
    p.nb = nb
    p.lo = lo
    p.hi = hi
    p.ty = ty
    p.ex = val
    return p
}

func freeLit(p *_Literal) {
    p.ex = nil
    _literalPool.Put(p)
}

func (self *_Literal) eval(m *[]byte) {
    var err error
    var val int64

    /* evaluate the expression */
    if val, err = self.ex.Evaluate(); err != nil {
        panic(err)
    }

    /* range check */
    if val < self.lo || val > self.hi {
        panic(fmt.Sprintf("expression out of range [%d, %d]: %d", self.lo, self.hi, val))
    }

    /* append to the output */
    p := (*[8]byte)(unsafe.Pointer(&val))
    *m = append(*m, p[:self.nb]...)
}

func (self *_Literal) Free() {
    self.ex.Free()
    freeLit(self)
}

func (self *_Literal) Type() PseudoType {
    return self.ty
}

func (self *_Literal) String() string {
    if self.ex.Type != expr.CONST {
        return fmt.Sprintf("%s ...", self.ty)
    } else {
        return fmt.Sprintf("%s %#x", self.ty, self.ex.Const)
    }
}

func (self *_Literal) Encode(m *[]byte, _ uintptr) int {
    if m != nil { self.eval(m) }
    return self.nb
}

// Byte represents a raw byte, encoded as is.
func Byte(v *expr.Expr) Pseudo {
    return newLit(PseudoByte, math.MinInt8, math.MaxUint8, 1, v)
}

// Word represents a raw uint16, encoded as little-endian bytes as is.
func Word(v *expr.Expr) Pseudo {
    return newLit(PseudoWord, math.MinInt16, math.MaxUint16, 2, v)
}

// Long represents a raw uint32, encoded as little-endian bytes as is.
func Long(v *expr.Expr) Pseudo {
    return newLit(PseudoLong, math.MinInt32, math.MaxUint32, 4, v)
}

// Quad represents a raw uint64, encoded as little-endian bytes as is.
func Quad(v *expr.Expr) Pseudo {
    return newLit(PseudoQuad, math.MinInt64, math.MaxInt64, 8, v)
}

type _Align struct {
    to uint64
    ex *expr.Expr
}

// Align instructs the assembler to align the PC to certain value, with optional padding characters.
func Align(to uint64, padding *expr.Expr) Pseudo {
    return &_Align {
        to: to,
        ex: padding,
    }
}

func (self *_Align) Free() {
    if self.ex != nil {
        self.ex.Free()
    }
}

func (self *_Align) Type() PseudoType {
    return PseudoAlign
}

func (self *_Align) String() string {
    if self.ex == nil {
        return fmt.Sprintf(".align %d", self.to)
    } else if self.ex.Type != expr.CONST {
        return fmt.Sprintf(".align %d, ...", self.to)
    } else {
        return fmt.Sprintf(".align %d, %#x", self.to, self.ex.Const)
    }
}

func (self *_Align) Encode(m *[]byte, pc uintptr) int {
    var err error
    var val int64

    /* check for alignment */
    if !ispow2(self.to) {
        panic(fmt.Errorf("alignment should be a power of 2, not %d", self.to))
    }

    /* evaluate the expression */
    if self.ex != nil {
        if val, err = self.ex.Evaluate(); err != nil {
            panic(err)
        }
    }

    /* verify fill char range */
    if val < 0 || val > math.MaxUint8 {
        panic(fmt.Errorf("alignment padding character out of range [0, 255]: %d", val))
    }

    /* calculate alignment size */
    size := align(int(pc), bits.TrailingZeros64(self.to))
    size -= int(pc)

    /* append alignment padding if needed */
    if m != nil { expandmm(m, size, uint8(val)) }
    return size
}
