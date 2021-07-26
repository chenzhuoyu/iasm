package x86_64

import (
    `fmt`
    `math`
    `math/bits`

    `github.com/chenzhuoyu/iasm/expr`
)

type (
	_PseudoType         int
    _InstructionEncoder func(*Program, ...interface{}) *Instruction
)

const (
    _PSEUDO_NOP _PseudoType = iota + 1
    _PSEUDO_BYTE
    _PSEUDO_WORD
    _PSEUDO_LONG
    _PSEUDO_QUAD
    _PSEUDO_DATA
    _PSEUDO_ALIGN
)

type _Pseudo struct {
    kind _PseudoType
    data []byte
    uint uint64
    expr *expr.Expr
}

func (self *_Pseudo) free() {
    switch self.kind {
        case _PSEUDO_BYTE  : fallthrough
        case _PSEUDO_WORD  : fallthrough
        case _PSEUDO_LONG  : fallthrough
        case _PSEUDO_QUAD  : fallthrough
        case _PSEUDO_ALIGN : self.expr.Free()
    }
}

func (self *_Pseudo) encode(m *[]byte, pc int) int {
    switch self.kind {
        case _PSEUDO_NOP   : return 0
        case _PSEUDO_BYTE  : self.encodeByte(m)      ; return 1
        case _PSEUDO_WORD  : self.encodeWord(m)      ; return 2
        case _PSEUDO_LONG  : self.encodeLong(m)      ; return 4
        case _PSEUDO_QUAD  : self.encodeQuad(m)      ; return 8
        case _PSEUDO_DATA  : self.encodeData(m)      ; return len(self.data)
        case _PSEUDO_ALIGN : self.encodeAlign(m, pc) ; return self.alignSize(pc)
        default            : panic("invalid pseudo instruction")
    }
}

func (self *_Pseudo) evalExpr(low int64, high uint64) int64 {
    if v, err := self.expr.Evaluate(); err != nil {
        panic(err)
    } else if v < low || uint64(v) > high {
        panic(fmt.Sprintf("expression out of range [%d, %d]: %d", low, high, v))
    } else {
        return v
    }
}

func (self *_Pseudo) alignSize(pc int) int {
    if !ispow2(self.uint) {
        panic(fmt.Sprintf("aligment should be a power of 2, not %d", self.uint))
    } else {
        return align(pc, bits.TrailingZeros64(self.uint)) - pc
    }
}

func (self *_Pseudo) encodeData(m *[]byte) {
    if m != nil {
        *m = append(*m, self.data...)
    }
}

func (self *_Pseudo) encodeByte(m *[]byte) {
    if m != nil {
        append8(m, byte(self.evalExpr(math.MinInt8, math.MaxUint8)))
    }
}

func (self *_Pseudo) encodeWord(m *[]byte) {
    if m != nil {
        append16(m, uint16(self.evalExpr(math.MinInt16, math.MaxUint16)))
    }
}

func (self *_Pseudo) encodeLong(m *[]byte) {
    if m != nil {
        append32(m, uint32(self.evalExpr(math.MinInt32, math.MaxUint32)))
    }
}

func (self *_Pseudo) encodeQuad(m *[]byte) {
    if m != nil {
        append64(m, uint64(self.evalExpr(math.MinInt64, math.MaxUint64)))
    }
}

func (self *_Pseudo) encodeAlign(m *[]byte, pc int) {
    if m != nil {
        if self.expr == nil {
            expandmm(m, self.alignSize(pc), 0)
        } else {
            expandmm(m, self.alignSize(pc), byte(self.evalExpr(math.MinInt8, math.MaxUint8)))
        }
    }
}

// Operands represents a sequence of operand required by an instruction.
type Operands [_MAX_ARGS]interface{}

// Instruction represents an unencoded instruction.
type Instruction struct {
    next   *Instruction
    pc     int
    nb     int
    len    int
    argc   int
    argv   [_MAX_ARGS]interface{}
    forms  [_MAX_FORMS]_Encoding
    pseudo _Pseudo
    branch bool
}

func (self *Instruction) add(flags int, encoder func(m *_Encoding, v []interface{})) {
    self.forms[self.len].flags = flags
    self.forms[self.len].encoder = encoder
    self.len++
}

func (self *Instruction) free() {
    self.clear()
    self.pseudo.free()
    freeInstruction(self)
}

func (self *Instruction) clear() {
    for i := 0; i < self.argc; i++ {
        freeLabel(self.argv[i])
        freeMemoryOperand(self.argv[i])
    }
}

func (self *Instruction) check(e *_Encoding) bool {
    if (e.flags & _REL1) != 0 {
        return isRel8(self.argv[0])
    } else if (e.flags & _REL4) != 0 {
        return isRel32(self.argv[0]) || isLabel(self.argv[0])
    } else {
        return true
    }
}

func (self *Instruction) encode(m *[]byte) int {
    n := math.MaxInt64
    p := (*_Encoding)(nil)

    /* check for pseudo-instructions */
    if self.pseudo.kind != 0 {
        self.nb = self.pseudo.encode(m, self.pc)
        return self.nb
    }

    /* find the shortest encoding */
    for i := 0; i < self.len; i++ {
        if e := &self.forms[i]; self.check(e) {
            if v := e.encode(self.argv[:self.argc]); v < n {
                n = v
                p = e
            }
        }
    }

    /* add to buffers if needed */
    if m != nil {
        *m = append(*m, p.bytes[:n]...)
    }

    /* update the instruction length */
    self.nb = n
    return n
}

// Program represents a sequence of instructions.
type Program struct {
    arch *Arch
    head *Instruction
    tail *Instruction
}

const (
    _NB_FAR  = 5    // far-branch takes 5 bytes to encode
    _NB_NEAR = 2    // near-branch (-128 ~ +127) takes 2 bytes to encode
    _NB_DIFF = _NB_FAR - _NB_NEAR
)

func (self *Program) clear() {
    for p := self.head; p != nil; p = p.next {
        p.free()
    }
}

func (self *Program) alloc(argc int, argv Operands) *Instruction {
    p := self.tail
    q := newInstruction(argc, argv)

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

func (self *Program) pseudo(kind _PseudoType) (p *Instruction) {
    p = self.alloc(0, Operands{})
    p.pseudo.kind = kind
    return
}

func (self *Program) require(isa ISA) {
    if !self.arch.HasISA(isa) {
        panic("ISA '" + isa.String() + "' was not enabled")
    }
}

/** Pseudo-Instructions **/

// Byte is a pseudo-instruction to add raw byte to the assembled code.
func (self *Program) Byte(v *expr.Expr) (p *Instruction) {
    p = self.pseudo(_PSEUDO_BYTE)
    p.pseudo.expr = v
    return
}

// Word is a pseudo-instruction to add raw uint16 as little-endian to the assembled code.
func (self *Program) Word(v *expr.Expr) (p *Instruction) {
    p = self.pseudo(_PSEUDO_WORD)
    p.pseudo.expr = v
    return
}

// Long is a pseudo-instruction to add raw uint32 as little-endian to the assembled code.
func (self *Program) Long(v *expr.Expr) (p *Instruction) {
    p = self.pseudo(_PSEUDO_LONG)
    p.pseudo.expr = v
    return
}

// Quad is a pseudo-instruction to add raw uint64 as little-endian to the assembled code.
func (self *Program) Quad(v *expr.Expr) (p *Instruction) {
    p = self.pseudo(_PSEUDO_QUAD)
    p.pseudo.expr = v
    return
}

// Data is a pseudo-instruction to add raw bytes to the assembled code.
func (self *Program) Data(v []byte) (p *Instruction) {
    p = self.pseudo(_PSEUDO_DATA)
    p.pseudo.data = v
    return
}

// Align is a pseudo-instruction to ensure the PC is aligned to a certain value.
func (self *Program) Align(align uint64, padding *expr.Expr) (p *Instruction) {
    p = self.pseudo(_PSEUDO_ALIGN)
    p.pseudo.uint = align
    p.pseudo.expr = padding
    return
}

/** Program Assembler **/

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
        p.Dest = self.pseudo(_PSEUDO_NOP)
    }
}

// Assemble assembles and links the entire program into machine code.
func (self *Program) Assemble(pc int) (ret []byte) {
    offs := 0
    next := true

    /* Pass 0: PC-precompute, assume all labeled branches are far-branches. */
    for p := self.head; p != nil; p = p.next {
        if p.pc = pc; p.branch && isLabel(p.argv[0]) {
            pc += _NB_FAR
        } else {
            pc += p.encode(nil)
        }
    }

    /* allocate space for the machine code */
    nb := pc
    ret = make([]byte, 0, nb)

    /* Pass 1: adjust all the jumps */
    for next {
        offs = 0
        next = false

        /* scan all the branches */
        for p := self.head; p != nil; p = p.next {
            var ok bool
            var lb *Label

            /* re-calculate the alignment here */
            if nb = p.nb; p.pseudo.kind == _PSEUDO_ALIGN {
                p.pc -= offs
                offs += nb - p.encode(nil)
                continue
            }

            /* only care about branches */
            if !p.branch {
                p.pc -= offs
                continue
            }

            /* check for labeled branches */
            if lb, ok = p.argv[0].(*Label); !ok {
                p.pc -= offs
                continue
            }

            /* calculate the jump offset */
            p.pc -= offs
            diff := lb.offset(p.pc, _NB_FAR)

            /* this is already a near jump */
            if p.nb == _NB_NEAR {
                continue
            }

            /* too far to be a near jump */
            if diff > 127 || diff < -128 {
                p.nb = _NB_FAR
                continue
            }

            /* a far jump becomes a near jump, calculate
             * the PC adjustment value and assemble again */
            next = true
            p.nb = _NB_NEAR
            offs += _NB_DIFF
        }
    }

    /* Pass 3: link all the cross references */
    for p := self.head; p != nil; p = p.next {
        for i := 0; i < p.argc; i++ {
            var ok bool
            var lb *Label
            var op *MemoryOperand

            /* resolve labels */
            if lb, ok = p.argv[i].(*Label); ok {
                p.argv[i] = lb.offset(p.pc, p.nb)
                continue
            }

            /* check for memory operands */
            if op, ok = p.argv[i].(*MemoryOperand); !ok {
                continue
            }

            /* check for label references */
            if op.Addr.Type != Reference {
                continue
            }

            /* replace the label with the real offset */
            op.Addr.Type = Offset
            op.Addr.Offset = op.Addr.Reference.offset(p.pc, p.nb)
        }
    }

    /* Pass 4: actually encode all the instructions */
    for p := self.head; p != nil; p = p.next {
        p.encode(&ret)
    }

    /* all done */
    return ret
}
