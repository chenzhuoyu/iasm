package x86_64

import (
    `encoding/binary`
    `math`
)

type _Node struct {
    next *Instruction
}

type _Pseudo struct {
    kind int
    data []byte
    uint uint64
}

const (
    _PSEUDO_NOP = iota + 1
    _PSEUDO_BYTE
    _PSEUDO_WORD
    _PSEUDO_LONG
    _PSEUDO_QUAD
    _PSEUDO_DATA
)

func (self *_Pseudo) encode(m *[]byte) int {
    switch self.kind {
        case _PSEUDO_NOP  : return 0
        case _PSEUDO_BYTE : self.encodeByte(m); return 1
        case _PSEUDO_WORD : self.encodeWord(m); return 2
        case _PSEUDO_LONG : self.encodeLong(m); return 4
        case _PSEUDO_QUAD : self.encodeQuad(m); return 8
        case _PSEUDO_DATA : self.encodeData(m); return len(self.data)
        default           : panic("invalid pseudo instruction")
    }
}

func (self *_Pseudo) encodeByte(m *[]byte) {
    if m != nil {
        *m = append(*m, byte(self.uint))
    }
}

func (self *_Pseudo) encodeWord(m *[]byte) {
    if m != nil {
        p := len(*m)
        *m = append(*m, 0, 0)
        binary.LittleEndian.PutUint16((*m)[p:], uint16(self.uint))
    }
}

func (self *_Pseudo) encodeLong(m *[]byte) {
    if m != nil {
        p := len(*m)
        *m = append(*m, 0, 0, 0, 0)
        binary.LittleEndian.PutUint32((*m)[p:], uint32(self.uint))
    }
}

func (self *_Pseudo) encodeQuad(m *[]byte) {
    if m != nil {
        p := len(*m)
        *m = append(*m, 0, 0, 0, 0, 0, 0, 0, 0)
        binary.LittleEndian.PutUint64((*m)[p:], self.uint)
    }
}

func (self *_Pseudo) encodeData(m *[]byte) {
    if m != nil {
        *m = append(*m, self.data...)
    }
}

// Operands represents a sequence of operand required by an instruction.
type Operands [_MAX_ARGS]interface{}

// Instruction represents an unencoded instruction.
type Instruction struct {
    _Node
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
        self.nb = self.pseudo.encode(m)
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

func (self *Program) pseudo(kind int) (p *Instruction) {
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
func (self *Program) Byte(v byte) (p *Instruction) {
    p = self.pseudo(_PSEUDO_BYTE)
    p.pseudo.uint = uint64(v)
    return
}

// Word is a pseudo-instruction to add raw uint16 as little-endian to the assembled code.
func (self *Program) Word(v uint16) (p *Instruction) {
    p = self.pseudo(_PSEUDO_WORD)
    p.pseudo.uint = uint64(v)
    return
}

// Long is a pseudo-instruction to add raw uint32 as little-endian to the assembled code.
func (self *Program) Long(v uint32) (p *Instruction) {
    p = self.pseudo(_PSEUDO_LONG)
    p.pseudo.uint = uint64(v)
    return
}

// Quad is a pseudo-instruction to add raw uint64 as little-endian to the assembled code.
func (self *Program) Quad(v uint64) (p *Instruction) {
    p = self.pseudo(_PSEUDO_QUAD)
    p.pseudo.uint = v
    return
}

// Data is a pseudo-instruction to add raw bytes to the assembled code.
func (self *Program) Data(v []byte) (p *Instruction) {
    p = self.pseudo(_PSEUDO_DATA)
    p.pseudo.data = v
    return
}

/** Program Assembler **/

// Free returns the Program object into pool.
// Any operation performed after Free is undefined behavior.
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
func (self *Program) Assemble() (ret []byte) {
    pc  := 0
    adj := 0
    asm := true

    /* Pass 0: PC-precompute, assume all labeled branches are far-branches. */
    for p := self.head; p != nil; p = p.next {
        if p.branch && isLabel(p.argv[0]) {
            p.pc, pc = pc, pc + _NB_FAR
        } else {
            p.pc, pc = pc, pc + p.encode(nil)
        }
    }

    /* allocate space for the machine code */
    nb := pc
    ret = make([]byte, 0, nb)

    /* Pass 1: adjust all the jumps */
    for asm {
        adj = 0
        asm = false

        /* scan all the branches */
        for p := self.head; p != nil; p = p.next {
            var ok bool
            var lb *Label

            /* only care about branches */
            if !p.branch {
                p.pc -= adj
                continue
            }

            /* check for labeled branches */
            if lb, ok = p.argv[0].(*Label); !ok {
                p.pc -= adj
                continue
            }

            /* calculate the jump offset */
            p.pc -= adj
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
            asm  = true
            adj += _NB_FAR - _NB_NEAR
            p.nb = _NB_NEAR
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
