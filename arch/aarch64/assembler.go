package aarch64

import (
    `fmt`
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

func (self *_AssemblerImpl) build(p *Program, line *asm.ParsedInstruction) error {
    if strings.ToUpper(line.Mnemonic) == "LDR" &&
       len(line.Operands)             == 2 &&
       line.Operands[0].Op            == asm.OpReg &&
       line.Operands[1].Op            == asm.OpSym &&
       line.Operands[1].Sym.Name      == _LitSym {
        return self.buildLdrLit(p, line.Operands[0].Reg, line.Operands[1].Sym.Value)
    } else {
        return self.buildEverythingElse(p, line)
    }
}

func (self *_AssemblerImpl) buildLdrLit(p *Program, reg asm.Register, lit interface{}) error {
    switch v := lit.(type) {
        default: {
            panic("unreachable")
        }

        /* integer immediates */
        case _LitInt: {
            p.LDI(reg, uint64(v))
            return nil
        }

        /* label references */
        case _LitName: {
            if lb, err := self.Symbols().Label(string(v)); err != nil {
                return err
            } else {
                p.ADRP(reg, lb)
                p.ADD(reg, reg, lb)
                return nil
            }
        }

        /* floating point immediates */
        case _LitFloat: {
            if !isFpImm8(v) {
                return fmt.Errorf("aarch64: %f is not encodable as an immediate value", v)
            } else {
                p.FMOV(reg, v)
                return nil
            }
        }
    }
}

func (self *_AssemblerImpl) buildEverythingElse(p *Program, line *asm.ParsedInstruction) error {
    var ok bool
    var vv []interface{}
    var fn _InstructionEncoder

    /* find the instruction */
    if fn, ok = Instructions[strings.ToUpper(line.Mnemonic)]; !ok {
        return fmt.Errorf("aarch64: unknown instruction: %s", line.Mnemonic)
    }

    /* convert the operands */
    for _, op := range line.Operands {
        switch op.Op {
            default: {
                panic("aarch64: parser yields an invalid operand kind")
            }

            /* simple values */
            case asm.OpImm   : vv = append(vv, op.Imm)
            case asm.OpReg   : vv = append(vv, op.Reg)
            case asm.OpMem   : vv = append(vv, Mem(op.Mem))
            case asm.OpMod   : vv = append(vv, op.Mod.Value)
            case asm.OpSym   : vv = append(vv, op.Sym.Value)
            case asm.OpPCrel : vv = append(vv, op.Rel)
            case asm.OpFpImm : vv = append(vv, op.FpImm)

            /* label references */
            case asm.OpLabel: {
                if tr, err := self.Symbols().Label(op.Label.Name); err != nil {
                    return err
                } else {
                    vv = append(vv, tr)
                }
            }
        }
    }

    /* encode the instruction */
    fn(p, vv...)
    return nil
}
