package x86_64

import (
    `math`
    `sync/atomic`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/expr`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Operands represents a sequence of operand required by an instruction.
type Operands [_N_args]interface{}

// InstructionDomain represents the domain of an instruction.
type InstructionDomain uint8

const (
    DomainGeneric InstructionDomain = iota
    DomainMMXSSE
    DomainAVX
    DomainFMA
    DomainCrypto
    DomainMask
    DomainAMDSpecific
    DomainMisc
    DomainPseudo
)

type (
    _BranchType uint8
)

const (
    _B_none _BranchType = iota
    _B_conditional
    _B_unconditional
)

// Instruction represents an unencoded instruction.
type Instruction struct {
    next   *Instruction
    refs   int64
    name   string
    pc     uintptr
    nb     int
    argc   int
    argv   Operands
    forms  [_N_forms]_Encoding
    nforms int
    pseudo asm.Pseudo
    branch _BranchType
    domain InstructionDomain
    prefix []byte
}

func (self *Instruction) add(flags int, encoder func(m *_Encoding, v []interface{})) {
    self.forms[self.nforms].flags = flags
    self.forms[self.nforms].encoder = encoder
    self.nforms++
}

func (self *Instruction) clear() {
    for i := 0; i < self.argc; i++ {
        if v, ok := self.argv[i].(tag.Disposable); ok {
            v.Free()
        }
    }
}

func (self *Instruction) check(e *_Encoding) bool {
    if (e.flags & _F_rel1) != 0 {
        return isRel8(self.argv[0])
    } else if (e.flags & _F_rel4) != 0 {
        return isRel32(self.argv[0]) || isLabel(self.argv[0])
    } else {
        return true
    }
}

func (self *Instruction) encode(m *[]byte) int {
    n := math.MaxInt64
    p := (*_Encoding)(nil)

    /* encode prefixes if any */
    if self.nb = len(self.prefix); m != nil {
        *m = append(*m, self.prefix...)
    }

    /* check for pseudo-instructions */
    if self.pseudo.Kind != 0 {
        self.nb += self.pseudo.Encode(m, self.pc)
        return self.nb
    }

    /* find the shortest encoding */
    for i := 0; i < self.nforms; i++ {
        if e := &self.forms[i]; self.check(e) {
            if v := e.encode(self.argv[:self.argc]); v < n {
                n = v
                p = e
            }
        }
    }

    /* add to buffer if needed */
    if m != nil {
        *m = append(*m, p.bytes[:n]...)
    }

    /* update the instruction length */
    self.nb += n
    return self.nb
}

func (self *Instruction) PC() uintptr {
    return self.pc
}

func (self *Instruction) Free() {
    if atomic.AddInt64(&self.refs, -1) == 0 {
        self.clear()
        self.pseudo.Free()
        freeInstruction(self)
    }
}

func (self *Instruction) Retain() asm.Instruction {
    atomic.AddInt64(&self.refs, 1)
    return self
}

func (self *Instruction) Domain() InstructionDomain {
    return self.domain
}

func (self *Instruction) Mnemonic() string {
    return self.name
}

func (self *Instruction) Operands() []interface{} {
    return self.argv[:self.argc]
}

/** Instruction Prefixes **/

const (
    _P_cs   = 0x2e
    _P_ds   = 0x3e
    _P_es   = 0x26
    _P_fs   = 0x64
    _P_gs   = 0x65
    _P_ss   = 0x36
    _P_lock = 0xf0
)

// CS overrides the memory operation of this instruction to CS.
func (self *Instruction) CS() *Instruction {
    self.prefix = append(self.prefix, _P_cs)
    return self
}

// DS overrides the memory operation of this instruction to DS,
// this is the default section for most instructions if not specified.
func (self *Instruction) DS() *Instruction {
    self.prefix = append(self.prefix, _P_ds)
    return self
}

// ES overrides the memory operation of this instruction to ES.
func (self *Instruction) ES() *Instruction {
    self.prefix = append(self.prefix, _P_es)
    return self
}

// FS overrides the memory operation of this instruction to FS.
func (self *Instruction) FS() *Instruction {
    self.prefix = append(self.prefix, _P_fs)
    return self
}

// GS overrides the memory operation of this instruction to GS.
func (self *Instruction) GS() *Instruction {
    self.prefix = append(self.prefix, _P_gs)
    return self
}

// SS overrides the memory operation of this instruction to SS.
func (self *Instruction) SS() *Instruction {
    self.prefix = append(self.prefix, _P_ss)
    return self
}

// LOCK causes the processor's LOCK# signal to be asserted during execution of
// the accompanying instruction (turns the instruction into an atomic instruction).
// In a multiprocessor environment, the LOCK# signal insures that the processor
// has exclusive use of any shared memory while the signal is asserted.
func (self *Instruction) LOCK() *Instruction {
    self.prefix = append(self.prefix, _P_lock)
    return self
}

// Program represents a sequence of x86_64 instructions.
type Program struct {
    arch *asm.Arch
    head *Instruction
    tail *Instruction
}

const (
    _N_near       = 2 // near-branch (-128 ~ +127) takes 2 bytes to encode
    _N_far_cond   = 6 // conditional far-branch takes 6 bytes to encode
    _N_far_uncond = 5 // unconditional far-branch takes 5 bytes to encode
)

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

func (self *Program) pseudo(kind asm.PseudoType) (p *Instruction) {
    p = self.alloc(kind.String(), 0, Operands{})
    p.domain = DomainPseudo
    p.pseudo.Kind = kind
    return
}

func (self *Program) require(isa ISA) {
    if !self.arch.HasFeature(isa) {
        panic("ISA '" + isa.String() + "' was not enabled")
    }
}

func (self *Program) branchSize(p *Instruction) int {
    switch p.branch {
        case _B_none          : panic("p is not a branch")
        case _B_conditional   : return _N_far_cond
        case _B_unconditional : return _N_far_uncond
        default               : panic("invalid instruction")
    }
}

func (self *Program) Free() {
    self.clear()
    freeProgram(self)
}

func (self *Program) Link(p *asm.Label) {
    if p.Dest != nil {
        panic("lable was alreay linked")
    } else {
        p.Dest = self.pseudo(asm.Nop).Retain()
    }
}

func (self *Program) Data(v []byte) asm.Instruction {
    p := self.pseudo(asm.Data)
    p.pseudo.Data = v
    return p
}

func (self *Program) Byte(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Byte)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Word(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Word)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Long(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Long)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Quad(v *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Quad)
    p.pseudo.Expr = v
    return p
}

func (self *Program) Align(align uint64, padding *expr.Expr) asm.Instruction {
    p := self.pseudo(asm.Align)
    p.pseudo.Uint = align
    p.pseudo.Expr = padding
    return p
}

func (self *Program) Assemble(pc uintptr) (ret []byte) {
    orig := pc
    next := true
    offs := uintptr(0)

    /* Pass 0: PC-precompute, assume all labeled branches are far-branches. */
    for p := self.head; p != nil; p = p.next {
        if p.pc = pc; !isLabel(p.argv[0]) || p.branch == _B_none {
            pc += uintptr(p.encode(nil))
        } else {
            pc += uintptr(self.branchSize(p))
        }
    }

    /* allocate space for the machine code */
    nb := int(pc - orig)
    ret = make([]byte, 0, nb)

    /* Pass 1: adjust all the jumps */
    for next {
        next = false
        offs = uintptr(0)

        /* scan all the branches */
        for p := self.head; p != nil; p = p.next {
            var ok bool
            var lb *asm.Label

            /* re-calculate the alignment here */
            if nb = p.nb; p.pseudo.Kind == asm.Align {
                p.pc -= offs
                offs += uintptr(nb - p.encode(nil))
                continue
            }

            /* adjust the program counter */
            p.pc -= offs
            lb, ok = p.argv[0].(*asm.Label)

            /* only care about labeled far-branches */
            if !ok || p.nb == _N_near || p.branch == _B_none {
                continue
            }

            /* calculate the jump offset */
            size := self.branchSize(p)
            diff := asm.ComputeOffset(lb, p.pc, size)

            /* too far to be a near jump */
            if diff > 127 || diff < -128 {
                p.nb = size
                continue
            }

            /* a far jump becomes a near jump, calculate
             * the PC adjustment value and assemble again */
            next = true
            p.nb = _N_near
            offs += uintptr(size - _N_near)
        }
    }

    /* Pass 3: link all the cross-references */
    for p := self.head; p != nil; p = p.next {
        for i := 0; i < p.argc; i++ {
            var ok bool
            var lb *asm.Label
            var op *asm.MemoryOperand

            /* resolve labels */
            if lb, ok = p.argv[i].(*asm.Label); ok {
                p.argv[i] = asm.ComputeOffset(lb, p.pc, p.nb)
                continue
            }

            /* resolve references in memory operands */
            if op, ok = p.argv[i].(*asm.MemoryOperand); ok {
                if lb, ok = op.Addr.(*asm.Label); ok {
                    op.Addr = asm.ComputeOffset(lb, p.pc, p.nb)
                }
            }
        }
    }

    /* Pass 4: actually encode all the instructions */
    for p := self.head; p != nil; p = p.next {
        p.encode(&ret)
    }

    /* all done */
    return ret
}
