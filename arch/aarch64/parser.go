package aarch64

import (
    `math`
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
)

type (
	_BranchTarget string
)

type _VecElem struct {
    reg _VReg
    fmt string
}

var _VectorModes = map[VecIndexMode]bool {
    ModeB: true,
    ModeH: true,
    ModeS: true,
    ModeD: true,
}

type _SymPair struct {
    sym string
    val interface{}
}

func mksym(sym string, val interface{}) (p _SymPair) {
    p.sym = sym
    p.val = val
    return
}

type _ParserImpl struct {
    lex       *asm.Tokenizer
    isMemCopy bool
}

func mkparser(lex *asm.Tokenizer) *_ParserImpl {
    return &_ParserImpl {
        lex       : lex,
        isMemCopy : false,
    }
}

func (self *_ParserImpl) err(pos int, msg string) *asm.SyntaxError {
    return &asm.SyntaxError {
        Pos    : pos,
        Row    : self.lex.Row,
        Src    : self.lex.Src,
        Reason : msg,
    }
}

func (self *_ParserImpl) eval() (ret int64) {
    ret = self.lex.Eval(nil)
    self.lex.ExpectPunc(asm.PuncRIndex)
    return
}

func (self *_ParserImpl) rela() asm.RelativeOffset {
    var iv int64
    var tk asm.Token

    /* record the tokenizer state */
    tp := self.lex.Pos
    tr := self.lex.Row

    /* a single dot */
    if tk = self.lex.Next(); tk.Ty == asm.TokenPunc {
        switch tk.Punc() {
            case asm.PuncPlus  : iv = 1
            case asm.PuncMinus : iv = -1
        }
    }

    /* neither "+" nor "-", it's a single dot, revert the tokenizer */
    if iv == 0 {
        self.lex.Pos = tp
        self.lex.Row = tr
        return 0
    }

    /* parse the offset absolute value */
    pos := self.lex.Pos
    val := self.lex.Value(nil) * iv

    /* range check */
    if val < math.MinInt32 || val > math.MaxInt32 {
        panic(self.err(pos, "relative offset out of range"))
    } else {
        return asm.RelativeOffset(val)
    }
}

func (self *_ParserImpl) elem(tk asm.Token) (ret _VecElem) {
    var sp int
    var ok bool

    /* the first token must be a name, and the next one must be a dot */
    if tk.Ty == asm.TokenName {
        self.lex.ExpectPuncNow(asm.PuncDot)
    } else {
        panic(self.err(tk.Pos, "vector register expected"))
    }

    /* must be vector registers */
    if ret.reg, ok = _VectorRegisters[strings.ToUpper(tk.Str)]; !ok {
        panic(self.err(tk.Pos, "invalid vector register name: " + tk.Str))
    }

    /* parse the vector arrangement */
    for sp = self.lex.Pos; sp < len(self.lex.Src); sp++ {
        if !isident(self.lex.Src[sp]) {
            break
        }
    }

    /* extract the arrangement */
    ret.fmt = strings.ToUpper(string(self.lex.Src[self.lex.Pos:sp]))
    self.lex.Pos = sp
    return
}

func (self *_ParserImpl) name(tk asm.Token) interface{} {
    ident := tk.Str
    upper := strings.ToUpper(ident)

    /* symbols, modifiers and trivial register names */
    if sym, ok := _Symbols[ident]       ; ok { return mksym(ident, sym) }
    if val, ok := _PStateTab[ident]     ; ok { return val }
    if mod, ok := _Modifiers[upper]     ; ok { return mod(self.modsize()) }
    if reg, ok := _SysRegisters[ident]  ; ok { return reg }
    if reg, ok := _CoreRegisters[upper] ; ok { return reg }
    if reg, ok := _SimdRegisters[upper] ; ok { return reg }

    /* if not a vector register, it must be a branch target */
    if _, ok := _VectorRegisters[upper]; !ok {
        return _BranchTarget(ident)
    }

    /* parse the register, then check if it's an indexed vector register */
    if v := self.elem(tk); !self.lex.MatchPuncNow(asm.PuncLIndex) {
        if vf, ok := _VecFormats[v.fmt]; !ok {
            panic(self.err(tk.Pos, "invalid arrangement specifier"))
        } else {
            return v.reg.toVec(vf)
        }
    } else {
        if im, ok := _VecIndexModes[v.fmt]; !ok {
            panic(self.err(self.lex.Pos, "invalid vector index mode"))
        } else if idx := self.eval(); idx < 0 || idx > int64(_MaxVecIndex[im]) {
            panic(self.err(self.lex.Pos, "vector index out of bounds"))
        } else {
            return v.reg.toIndex(im, uint8(idx))
        }
    }
}

func (self *_ParserImpl) vector() asm.Register {
    var nb int
    var vw _VecElem
    var vx _VecElem

    /* parse the first element */
    ps := self.lex.Pos
    v0 := self.elem(self.lex.Next())

    /* parse all the registers, should be at least one */
    for nb, vw = 1, v0; self.lex.MatchPunc(asm.PuncComma); nb++ {
        tk := self.lex.Next()
        vx, vw = vw, self.elem(tk)

        /* check for vector size and format */
        if nb >= 4 {
            panic(self.err(tk.Pos, "too may vector elements"))
        } else if vw.fmt != vx.fmt {
            panic(self.err(tk.Pos, "vector format must be identical for each element"))
        } else if vw.reg != (vx.reg + 1) % 32 {
            panic(self.err(tk.Pos, "vector elements must be consecutive registers"))
        }
    }

    /* must end with a "}", and if then found an index bracket, it's a single structure vector */
    if self.lex.ExpectPunc(asm.PuncRCurly); !self.lex.MatchPunc(asm.PuncLIndex) {
        if vf, ok := _VecFormats[v0.fmt]; !ok {
            panic(self.err(ps, "invalid vector format"))
        } else {
            return VecN(v0.reg.toVec(vf), nb)
        }
    } else {
        if im, ok := _VecIndexModes[v0.fmt]; !ok || !_VectorModes[im] {
            panic(self.err(ps, "invalid indexed vector format"))
        } else if idx := self.eval(); idx < 0 || idx > int64(_MaxVecIndex[im]) {
            panic(self.err(self.lex.Pos, "vector index out of bounds"))
        } else {
            return IndexN(v0.reg.toIndex(im, 0), nb, uint8(idx))
        }
    }
}

func (self *_ParserImpl) memory() (mem asm.MemoryAddress) {
    var ok bool
    var tk asm.Token
    var rr asm.Register

    /* must be a register */
    if tk = self.lex.Next(); tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "register expected"))
    }

    /* find the register */
    if mem.Base, ok = _CoreRegisters[strings.ToUpper(tk.Str)]; !ok {
        panic(self.err(tk.Pos, "register expected"))
    }

    /* base register must be Xr for CPY* instructions, or may also be SP for any other cases */
    if self.isMemCopy {
        if !isXr(mem.Base) {
            panic(self.err(tk.Pos, "X registers expected"))
        }
    } else {
        if !isXrOrSP(mem.Base) {
            panic(self.err(tk.Pos, "X registers or SP expected"))
        }
    }

    /* "[<Xr|SP>]!" (special-case syntax requirement for CPY* instructions) */
    if self.isMemCopy {
        mem.Ext = PreIndex
        self.lex.ExpectPunc(asm.PuncRIndex)
        self.lex.ExpectPunc(asm.PuncBang)
        return mem
    }

    /* "[<Xr|SP>]" or "[<Xr|SP>], #imm" */
    if !self.lex.MatchPunc(asm.PuncComma) {
        self.lex.ExpectPunc(asm.PuncRIndex)
        return self.postindex(mem)
    }

    /* "[<Xr|SP>, #imm]" or "[<Xr|SP>, #imm]!" */
    if tk = self.lex.Next(); tk.IsPunc(asm.PuncHash) {
        val := self.lex.Value(nil)
        self.lex.ExpectPunc(asm.PuncRIndex)
        return self.preindex(mem, val)
    }

    /* must be an identifier (register name) */
    if tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "index register expected"))
    }

    /* this identifier must be a register name */
    if rr, ok = _CoreRegisters[strings.ToUpper(tk.Str)]; !ok {
        panic(self.err(tk.Pos, "index register expected"))
    }

    /* may have extension or shift modifier, otherwise it's a basic memory operand */
    if !self.lex.MatchPunc(asm.PuncComma) {
        mem.Ext = Basic
    } else {
        mem.Ext = self.modifier()
    }

    /* match the closing "]" */
    mem.Index = rr
    self.lex.ExpectPunc(asm.PuncRIndex)

    /* guard for the invalid pre-indexing */
    if self.lex.MatchPunc(asm.PuncBang) {
        panic(self.err(self.lex.Pos, "pre-indexing conflits with register indexing"))
    } else {
        return mem
    }
}

func (self *_ParserImpl) modsize() uint8 {
    for {
        tp := self.lex.Pos
        tr := self.lex.Row
        tk := self.lex.Read()

        /* check token types */
        switch tk.Ty {
            default: {
                self.lex.Pos = tp
                self.lex.Row = tr
                return 0
            }

            /* simple cases */
            case asm.TokenEnd   : return 0
            case asm.TokenInt   : return uint8(tk.Uint)
            case asm.TokenFp64  : panic(self.err(tk.Pos, "extension or shift amount must be integers"))
            case asm.TokenSpace : break

            /* maybe unary operators, need further inspections */
            case asm.TokenPunc: {
                self.lex.Pos = tp
                self.lex.Row = tr

                /* only a handful of operators can act as value prefixes */
                switch tk.Punc() {
                    case asm.PuncPlus   : break
                    case asm.PuncMinus  : break
                    case asm.PuncTilde  : break
                    case asm.PuncLBrace : break
                    default             : return 0
                }

                /* expression must be non-negative */
                if n := self.lex.Value(nil); n < 0 {
                    panic(self.err(tk.Pos, "extension or shift amount must be non-negative"))
                } else {
                    return uint8(n)
                }
            }
        }
    }
}

func (self *_ParserImpl) modifier() Modifier {
    pos := self.lex.Pos
    row := self.lex.Row
    tok := self.lex.Next()

    /* should be an identifier, otherwise it's definately not a modifier */
    if tok.Ty != asm.TokenName {
        self.lex.Pos = pos
        self.lex.Row = row
        return nil
    }

    /* construct the modifier */
    if mod, ok := _Modifiers[strings.ToUpper(tok.Str)]; !ok {
        panic(self.err(tok.Pos, "invalid modifier"))
    } else {
        return mod(self.modsize())
    }
}

func (self *_ParserImpl) preindex(mem asm.MemoryAddress, offs int64) asm.MemoryAddress {
    if !self.lex.MatchPunc(asm.PuncBang) {
        mem.Ext = Basic
        mem.Offset = int32(offs & 0xfff)
        return mem
    } else {
        mem.Ext = PreIndex
        mem.Offset = int32(offs & 0x1ff)
        return mem
    }
}

func (self *_ParserImpl) postindex(mem asm.MemoryAddress) asm.MemoryAddress {
    pos := self.lex.Pos
    row := self.lex.Row
    tok := self.lex.Next()

    /* pre-indexing without offset is only allowed in CPY* instructions */
    if tok.IsPunc(asm.PuncBang) {
        panic(self.err(tok.Pos, "pre-indexing without offset is only allowed in CPY* instructions"))
    }

    /* post-indexing */
    if tok.IsPunc(asm.PuncComma) {
        if self.lex.Next().IsPunc(asm.PuncHash) {
            mem.Ext = PostIndex
            mem.Offset = int32(self.lex.Value(nil) & 0xfff)
            return mem
        }
    }

    /* does not match either of above, it's a basic memory address */
    mem.Ext = Basic
    self.lex.Pos = pos
    self.lex.Row = row
    return mem
}

func (self *_ParserImpl) value(ins *asm.ParsedInstruction) {
    switch tk := self.lex.Next(); {
        default: {
            panic(self.err(tk.Pos, "invalid operand for instruction " + ins.Mnemonic))
        }

        /* trailing comma, that's an error */
        case tk.Ty == asm.TokenEnd: {
            panic(self.err(tk.Pos, "trailing comma is not allowed"))
        }

        /* special case register for CPY* instructions */
        case tk.Ty == asm.TokenName && self.isMemCopy: {
            if reg, ok := self.name(tk).(asm.Register); !ok {
                panic(self.err(tk.Pos, "register expected"))
            } else {
                self.lex.ExpectPunc(asm.PuncBang)
                ins.Reg(reg)
            }
        }

        /* normal registers or symbols */
        case tk.Ty == asm.TokenName && !self.isMemCopy: {
            switch v := self.name(tk).(type) {
                case asm.Register  : ins.Reg(v)
                case Modifier      : ins.Mod(v)
                case _SymPair      : ins.Sym(v.sym, v.val)
                case _BranchTarget : ins.Target(string(v))
                default            : panic("unreachable")
            }
        }

        /* relative address */
        case tk.IsPunc(asm.PuncDot) && !self.isMemCopy: {
            ins.PCrel(self.rela())
        }

        /* immediate values */
        case tk.IsPunc(asm.PuncHash) && !self.isMemCopy: {
            switch tk = self.lex.Read(); tk.Ty {
                case asm.TokenInt  : ins.Imm(int64(tk.Uint))
                case asm.TokenFp64 : ins.FpImm(tk.Fp64)
                default            : panic(self.err(tk.Pos, "immediate value expected"))
            }
        }

        /* label references */
        case tk.IsPunc(asm.PuncEqual) && !self.isMemCopy: {
            switch tk = self.lex.Next(); tk.Ty {
                default: {
                    panic(self.err(tk.Pos, "identifier or const expression expected"))
                }

                /* LDR Rd, =<const or label or var-name> */
                case asm.TokenInt  : ins.Sym(_LitSym, _LitInt(tk.Uint))
                case asm.TokenName : ins.Sym(_LitSym, _LitName(tk.Str))
                case asm.TokenFp64 : ins.Sym(_LitSym, _LitFloat(tk.Fp64))

                /* LDR Rd, =(<expr>) */
                case asm.TokenPunc: {
                    if tk.Punc() != asm.PuncLBrace {
                        panic(self.err(tk.Pos, "identifier or const expression expected"))
                    } else {
                        ins.Imm(self.lex.Value(nil))
                        self.lex.ExpectPunc(asm.PuncRBrace)
                    }
                }
            }
        }

        /* register vectors */
        case tk.IsPunc(asm.PuncLCurly) && !self.isMemCopy: {
            ins.Reg(self.vector())
        }

        /* memory operands */
        case tk.IsPunc(asm.PuncLIndex): {
            ins.Mem(self.memory())
        }
    }
}

func (self *_ParserImpl) parse(ins *asm.ParsedInstruction) {
    tk := self.lex.Next()
    str := tk.Str

    /* must be a name */
    if tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "identifier expected"))
    }

    /* mnemonic is case-insensitive */
    ins.Mnemonic = strings.ToUpper(str)
    self.isMemCopy = strings.HasPrefix(ins.Mnemonic, "CPY")

    /* check if the instruction has operands */
    if self.lex.MatchEnd() {
        return
    }

    /* construct a placeholder token */
    tk.Ty = asm.TokenPunc
    tk.Uint = uint64(asm.PuncComma)

    /* parse all the operands */
    for tk.IsPunc(asm.PuncComma) {
        self.value(ins)
        tk = self.lex.Next()
    }

    /* should be the end of instruction */
    if tk.Ty != asm.TokenEnd {
        panic(self.err(tk.Pos, "garbage after instruction"))
    }
}
