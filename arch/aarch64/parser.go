package aarch64

import (
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/expr`
)

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

func (self *_ParserImpl) name(tk asm.Token) (string, asm.Register) {
    var ok bool
    var rx asm.Register

    /* normal registers are case-insensitive */
    pos := 0
    ivr := false
    idx := int64(0)
    key := strings.ToUpper(tk.Str)

    /* symbols, case-sensitive */
    if _, ok = _Symbols[tk.Str]; ok {
        return tk.Str, nil
    }

    /* core registers */
    if rx, ok = _Registers[key]; ok {
        return "", rx
    }

    /* system registers, case sensitive */
    if rx, ok = _SysRegisters[tk.Str]; ok {
        return "", rx
    }

    /* must be vector registers */
    if rx, ok = _VecRegisters[key]; !ok {
        panic(self.err(tk.Pos, "invalid register or symbol name: " + tk.Str))
    }

    /* must be a dot */
    if tk = self.lex.Read(); tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncDot {
        panic(self.err(tk.Pos, `"." expected`))
    }

    /* parse the vector arrangement */
    for pos = self.lex.Pos; pos < len(self.lex.Src); pos++ {
        if !isident(self.lex.Src[pos]) {
            break
        }
    }

    /* extract the arrangement */
    row := self.lex.Row
    pfx := strings.ToUpper(string(self.lex.Src[self.lex.Pos:pos]))
    self.lex.Pos = pos

    /* check for indexed vector register */
    if tx := self.lex.Read(); tx.Ty == asm.TokenPunc && tx.Punc() == asm.PuncLIndex {
        ivr = true
        idx = self.eval(self.lex.Pos)
    }

    /* indexed vector register */
    if ivr {
        if idx < 0 || idx > MaxVecIndex {
            panic(self.err(self.lex.Pos, "vector index out of bounds"))
        } else if im, ok := _VecIndexModes[pfx]; !ok {
            panic(self.err(pos, "invalid vector index mode"))
        } else {
            return "", rx.(_VReg).toIndex(im, uint8(idx))
        }
    }

    /* restore lexer state if not an indexed vector register */
    self.lex.Pos = pos
    self.lex.Row = row

    /* also, it must have the size prefix */
    if pfx == "" {
        panic(self.err(tk.Pos, "invalid arrangement specifier: " + tk.Str))
    } else if vf, ok := _VecFormats[pfx]; !ok {
        panic(self.err(tk.Pos, "invalid arrangement specifier"))
    } else {
        return "", rx.(_VReg).toVec(vf)
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
                if sym, reg := self.name(tk); reg != nil {
                    ins.Reg(reg)
                } else {
                    ins.Sym(sym)
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

            /* memory operands */
            case tt == asm.TokenPunc && tk.Punc() == asm.PuncLBrace: {
                panic("iasm: not implemented") // TODO: this
            }
        }
    }
}
