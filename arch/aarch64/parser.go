package aarch64

import (
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/expr`
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

type _ParserImpl struct {
    lex *asm.Tokenizer
}

func mkparser(lex *asm.Tokenizer) *_ParserImpl {
    return &_ParserImpl { lex }
}

func (self *_ParserImpl) err(pos int, msg string) *asm.SyntaxError {
    return &asm.SyntaxError {
        Pos    : pos,
        Row    : self.lex.Row,
        Src    : self.lex.Src,
        Reason : msg,
    }
}

func (self *_ParserImpl) eval(p int) int64 {
    var r int64
    var e error

    /* searching start */
    n := 1
    q := p + 1

    /* find the end of expression */
    for n > 0 && q < len(self.lex.Src) {
        switch self.lex.Src[q] {
            case '[' : q++; n++
            case ']' : q++; n--
            default  : q++
        }
    }

    /* check for EOF */
    if n != 0 {
        panic(self.err(q, "unexpected EOF when parsing expressions"))
    }

    /* evaluate the expression */
    if r, e = expr.Eval(string(self.lex.Src[p:q - 1]), nil); e != nil {
        panic(self.err(p, "cannot evaluate expression: " + e.Error()))
    }

    /* skip the last "]" */
    self.lex.Pos = q
    return r
}

func (self *_ParserImpl) delim(dp asm.Punctuation) bool {
    pos := self.lex.Pos
    row := self.lex.Row
    tok := self.lex.Next()

    /* check for the punctuation */
    if tok.Ty == asm.TokenPunc && tok.Punc() == dp {
        return true
    }

    /* otherwise rewind the tokenizer */
    self.lex.Pos = pos
    self.lex.Row = row
    return false
}

func (self *_ParserImpl) elem(tk asm.Token) (ret _VecElem) {
    var sp int
    var ok bool

    /* the first token must be a name */
    if tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "vector register expected"))
    }

    /* must be vector registers */
    if ret.reg, ok = _VecRegisters[strings.ToUpper(tk.Str)]; !ok {
        panic(self.err(tk.Pos, "invalid vector register name: " + tk.Str))
    }

    /* must be a dot */
    if tk = self.lex.Read(); tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncDot {
        panic(self.err(tk.Pos, `"." expected`))
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

func (self *_ParserImpl) name(tk asm.Token) (asm.Register, string, interface{}) {
    var ok bool
    var sv interface{}
    var rx asm.Register

    /* the first token must be a name */
    if tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "symbol or register expected"))
    }

    /* symbols, case-sensitive */
    if sv, ok = _Symbols[tk.Str]; ok {
        return nil, tk.Str, sv
    }

    /* core registers */
    if rx, ok = _Registers[strings.ToUpper(tk.Str)]; ok {
        return rx, "", nil
    }

    /* pstate fields, case-sensitive */
    if rx, ok = _PStateTab[tk.Str]; ok {
        return rx, "", nil
    }

    /* system registers, case sensitive */
    if rx, ok = _SysRegisters[tk.Str]; ok {
        return rx, "", nil
    }

    /* must be vector registers */
    ivr := false
    idx := int64(0)
    reg := self.elem(tk)
    pos := self.lex.Pos
    row := self.lex.Row

    /* check for indexed vector register */
    if tx := self.lex.Read(); tx.Ty == asm.TokenPunc && tx.Punc() == asm.PuncLIndex {
        ivr = true
        idx = self.eval(self.lex.Pos)
    }

    /* indexed vector register */
    if ivr {
        if im, ok := _VecIndexModes[reg.fmt]; !ok {
            panic(self.err(pos, "invalid vector index mode"))
        } else if idx < 0 || idx > int64(_MaxVecIndex[im]) {
            panic(self.err(self.lex.Pos, "vector index out of bounds"))
        } else {
            return reg.reg.toIndex(im, uint8(idx)), "", nil
        }
    }

    /* restore lexer state if not an indexed vector register */
    self.lex.Pos = pos
    self.lex.Row = row

    /* also, it must have the size prefix */
    if reg.fmt == "" {
        panic(self.err(tk.Pos, "invalid arrangement specifier: " + tk.Str))
    } else if vf, ok := _VecFormats[reg.fmt]; !ok {
        panic(self.err(tk.Pos, "invalid arrangement specifier"))
    } else {
        return reg.reg.toVec(vf), "", nil
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
    for nb, vw = 1, v0; self.delim(asm.PuncComma); nb++ {
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

    /* must end with a "}" */
    if tk := self.lex.Next(); tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncRCurly {
        panic(self.err(tk.Pos, "'}' expected"))
    }

    /* save the tokenizer cursor */
    pos := self.lex.Pos
    row := self.lex.Row

    /* found an index bracket, it's a single structure vector */
    if tk := self.lex.Next(); tk.Ty == asm.TokenPunc && tk.Punc() == asm.PuncLIndex {
        if im, ok := _VecIndexModes[v0.fmt]; !ok || !_VectorModes[im] {
            panic(self.err(ps, "invalid indexed vector format"))
        } else if idx := self.eval(self.lex.Pos); idx < 0 || idx > int64(_MaxVecIndex[im]) {
            panic(self.err(self.lex.Pos, "vector index out of bounds"))
        } else {
            return IndexN(v0.reg.toIndex(im, 0), nb, uint8(idx))
        }
    }

    /* otherwise it's a multi-structure vector, revert the tokenizer */
    self.lex.Pos = pos
    self.lex.Row = row

    /* convert the vector format */
    if vf, ok := _VecFormats[v0.fmt]; !ok {
        panic(self.err(ps, "invalid vector format"))
    } else {
        return VecN(v0.reg.toVec(vf), nb)
    }
}

func (self *_ParserImpl) parse(ins *asm.ParsedInstruction) {
    ff := true
    tt := asm.TokenEnd
    tk := self.lex.Next()

    /* must be an instruction */
    if tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "identifier expected"))
    } else {
        ins.Mnemonic = strings.ToLower(tk.Str)
    }

    /* parse all the operands */
    for {
        tk = self.lex.Next()
        tt = tk.Ty

        /* check for end of line */
        if tt == asm.TokenEnd {
            break
        }

        /* expect a comma if not the first operand */
        if !ff {
            if tt == asm.TokenPunc && tk.Punc() == asm.PuncComma {
                tk = self.lex.Next()
            } else {
                panic(self.err(tk.Pos, "',' expected"))
            }
        }

        /* not the first operand anymore */
        ff = false
        tt = tk.Ty

        /* parse the operand */
        switch {
            default: {
                panic(self.err(tk.Pos, "invalid operand"))
            }

            /* registers or symbols */
            case tt == asm.TokenName: {
                if reg, name, value := self.name(tk); reg != nil {
                    ins.Reg(reg)
                } else {
                    ins.Sym(name, value)
                }
            }

            /* literal values */
            case tt == asm.TokenPunc && tk.Punc() == asm.PuncHash: {
                switch tk = self.lex.Next(); tk.Ty {
                    case asm.TokenInt  : ins.Imm(int64(tk.Uint))
                    case asm.TokenFp64 : ins.FpImm(tk.Fp64)
                    default            : panic(self.err(tk.Pos, "literal value expected"))
                }
            }

            /* register vectors */
            case tt == asm.TokenPunc && tk.Punc() == asm.PuncLCurly: {
                ins.Reg(self.vector())
            }

            /* memory operands */
            case tt == asm.TokenPunc && tk.Punc() == asm.PuncLIndex: {
                panic("iasm: not implemented") // TODO: this
            }
        }
    }
}
