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
            case asm.OpImm   : ops = append(ops, op.Imm)
            case asm.OpReg   : ops = append(ops, op.Reg)
            case asm.OpMem   : self.buildMem(&ops, op.Mem)
            case asm.OpLabel : self.buildLabel(&ops, op.Label)
            default          : panic("parser yields an invalid operand kind")
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

func (self *_AssemblerImpl) buildMem(ops *[]interface{}, addr asm.MemoryAddress) {
    if addr.Base != rip {
        *ops = append(*ops, Mem(addr))
    } else {
        *ops = append(*ops, Mem(asm.RelativeOffset(addr.Offset)))
    }
}

func (self *_AssemblerImpl) buildLabel(ops *[]interface{}, label asm.ParsedLabel) {
    if tr, err := self.Symbols().Label(label.Name); err != nil {
        panic(err)
    } else if label.Kind == asm.BranchTarget {
        *ops = append(*ops, tr)
    } else {
        *ops = append(*ops, Mem(tr))
    }
}
