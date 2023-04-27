package x86_64

import (
    `fmt`
    `strconv`
    `strings`
    `unsafe`

    `github.com/chenzhuoyu/iasm/asm`
)

type _AssemblerImpl struct {
    asm.Assembler
}

func asmthis(p *asm.Assembler) *_AssemblerImpl {
    return (*_AssemblerImpl)(unsafe.Pointer(p))
}

func (self *_AssemblerImpl) build(p *Program, line *asm.ParsedInstruction) (err error) {
    var ok  bool
    var pfx []byte
    var ops []interface{}
    var enc _InstructionEncoder

    /* convert to lower-case */
    opts := self.Options()
    name := strings.ToLower(line.Mnemonic)

    /* fix register-addressing branches if needed */
    if opts.PermissiveMnemonic && len(line.Operands) == 1 {
        switch {
            case name == "retq"                                        : name = "ret"
            case name == "movabsq"                                     : name = "movq"
            case name == "jmp"   && line.Operands[0].Op != asm.OpLabel : name = "jmpq"
            case name == "jmpq"  && line.Operands[0].Op == asm.OpLabel : name = "jmp"
            case name == "call"  && line.Operands[0].Op != asm.OpLabel : name = "callq"
            case name == "callq" && line.Operands[0].Op == asm.OpLabel : name = "call"
        }
    }

    /* lookup from the alias table if needed */
    if opts.PermissiveMnemonic {
        enc, ok = _InstructionAliases[name]
    }

    /* lookup from the instruction table */
    if !ok {
        enc, ok = Instructions[name]
    }

    /* remove size suffix if possible */
    if !ok && opts.PermissiveMnemonic {
        switch i := len(name) - 1; name[i] {
            case 'b', 'w', 'l', 'q': {
                enc, ok = Instructions[name[:i]]
            }
        }
    }

    /* check for instruction name */
    if !ok {
        return self.Error("no such instruction: " + strconv.Quote(name))
    }

    /* allocate memory for prefix if any */
    if len(line.Prefixes) != 0 {
        pfx = make([]byte, len(line.Prefixes))
    }

    /* convert the prefixes */
    for i, v := range line.Prefixes {
        switch v {
            case PrefixLock      : pfx[i] = _P_lock
            case PrefixSegmentCS : pfx[i] = _P_cs
            case PrefixSegmentDS : pfx[i] = _P_ds
            case PrefixSegmentES : pfx[i] = _P_es
            case PrefixSegmentFS : pfx[i] = _P_fs
            case PrefixSegmentGS : pfx[i] = _P_gs
            case PrefixSegmentSS : pfx[i] = _P_ss
            default              : panic("unreachable: invalid segment prefix")
        }
    }

    /* convert the operands */
    for _, op := range line.Operands {
        switch op.Op {
            default: {
                panic("parser yields an invalid operand kind")
            }

            /* simple values */
            case asm.OpImm   : ops = append(ops, op.Imm)
            case asm.OpReg   : ops = append(ops, op.Reg)
            case asm.OpPCrel : ops = append(ops, op.Rel)

            /* memory operands */
            case asm.OpMem: {
                if op.Mem.Base != rip {
                    ops = append(ops, Mem(op.Mem))
                } else {
                    ops = append(ops, Mem(asm.RelativeOffset(op.Mem.Offset)))
                }
            }

            /* label references */
            case asm.OpLabel: {
                if tr, err := self.Symbols().Label(op.Label.Name); err != nil {
                    panic(err)
                } else if op.Label.Kind == asm.BranchTarget {
                    ops = append(ops, tr)
                } else {
                    ops = append(ops, Mem(tr))
                }
            }
        }
    }

    /* catch any exceptions in the encoder */
    defer func() {
        if v := recover(); v != nil {
            err = self.Error(fmt.Sprint(v))
        }
    }()

    /* encode the instruction */
    enc(p, ops...).prefix = pfx
    return nil
}
