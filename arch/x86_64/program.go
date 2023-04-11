package x86_64

import (
    `math`

    `github.com/chenzhuoyu/iasm/asm`
)

// Instruction represents an unencoded instruction.
type Instruction struct {
    asm.Instruction
    nb     int
    forms  [_N_forms]_Encoding
    nforms int
    prefix []byte
}

func (self *Instruction) add(flags int, encoder func(m *_Encoding, v []interface{})) {
    self.forms[self.nforms].flags = flags
    self.forms[self.nforms].encoder = encoder
    self.nforms++
}

func (self *Instruction) check(e *_Encoding) bool {
    if (e.flags & _F_rel1) != 0 {
        return isRel8(self.Argv[0])
    } else if (e.flags & _F_rel4) != 0 {
        return isRel32(self.Argv[0]) || isLabel(self.Argv[0])
    } else {
        return true
    }
}

func (self *Instruction) encode(m *[]byte) int {
    n := math.MaxInt64
    p := (*_Encoding)(nil)

    /* check for pseudo-instructions */
    if self.Pseudo.Kind != 0 {
        self.nb = self.Pseudo.Encode(m, self.PC)
        return self.nb
    }

    /* encode prefixes if any */
    if self.nb = len(self.prefix); m != nil {
        *m = append(*m, self.prefix...)
    }

    /* find the shortest encoding */
    for i := 0; i < self.nforms; i++ {
        if e := &self.forms[i]; self.check(e) {
            if v := e.encode(self.Argv[:self.Argc]); v < n {
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
    asm.Program
}

const (
    _N_near       = 2 // near-branch (-128 ~ +127) takes 2 bytes to encode
    _N_far_cond   = 6 // conditional far-branch takes 6 bytes to encode
    _N_far_uncond = 5 // unconditional far-branch takes 5 bytes to encode
)

func (self *Program) alloc(name string, argc int, argv asm.Operands) *Instruction {
    return this(self.Append(name, argc, argv))
}

func (self *Program) encode(m *[]byte) {
    for _, p := range self.Instr {
        this(p).encode(m)
    }
}

func (self *Program) assemble(pc uintptr) (ret []byte) {
    orig := pc
    next := true
    offs := uintptr(0)

    /* Pass 0: PC-precompute, assume all labeled branches are far-branches. */
    for _, p := range self.Instr {
        if p.PC = pc; isLabel(p.Argv[0]) && p.Branch != asm.BranchNone {
            pc += uintptr(brsize(p))
        } else {
            pc += uintptr(this(p).encode(nil))
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
        for _, m := range self.Instr {
            p := this(m)
            k := p.Pseudo.Kind

            /* re-calculate the alignment here */
            if k == asm.Align {
                p.PC -= offs
                offs += uintptr(p.nb - p.encode(nil))
                continue
            }

            /* adjust the program counter */
            p.PC -= offs
            lb, ok := p.Argv[0].(*asm.Label)

            /* only care about labeled far-branches */
            if !ok || p.nb == _N_near || p.Branch == asm.BranchNone {
                continue
            }

            /* calculate the jump offset */
            size := brsize(m)
            diff := asm.ComputeOffset(lb, p.PC, size)

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
    for _, p := range self.Instr {
        for i := 0; i < p.Argc; i++ {
            var ok bool
            var lb *asm.Label
            var op *asm.MemoryOperand

            /* resolve labels */
            if lb, ok = p.Argv[i].(*asm.Label); ok {
                p.Argv[i] = asm.ComputeOffset(lb, p.PC, this(p).nb)
                continue
            }

            /* resolve references in memory operands */
            if op, ok = p.Argv[i].(*asm.MemoryOperand); ok {
                if lb, ok = op.Addr.(*asm.Label); ok {
                    op.Addr = asm.ComputeOffset(lb, p.PC, this(p).nb)
                }
            }
        }
    }

    /* Pass 4: actually encode all the instructions */
    self.encode(&ret)
    return ret
}
