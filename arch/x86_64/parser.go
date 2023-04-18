package x86_64

import (
    `fmt`
    `math`
    `strconv`
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/expr`
)

type (
	_PrefixImpl byte
)

const (
    PrefixLock _PrefixImpl = iota
    PrefixSegmentCS
    PrefixSegmentDS
    PrefixSegmentES
    PrefixSegmentFS
    PrefixSegmentGS
    PrefixSegmentSS
)

func (self _PrefixImpl) String() string {
    switch self {
        case PrefixLock      : return "LOCK"
        case PrefixSegmentCS : return "CS"
        case PrefixSegmentDS : return "DS"
        case PrefixSegmentES : return "ES"
        case PrefixSegmentFS : return "FS"
        case PrefixSegmentGS : return "GS"
        case PrefixSegmentSS : return "SS"
        default              : return "???"
    }
}

func (self _PrefixImpl) EncodePrefix(p *asm.Instruction) {
    switch self {
        case PrefixLock      : this(p).LOCK()
        case PrefixSegmentCS : this(p).CS()
        case PrefixSegmentDS : this(p).DS()
        case PrefixSegmentES : this(p).ES()
        case PrefixSegmentFS : this(p).FS()
        case PrefixSegmentGS : this(p).GS()
        case PrefixSegmentSS : this(p).SS()
        default              : panic("unreachable")
    }
}

const (
    rip Register64 = 0xff
)

var _RegBranch = map[string]bool {
    "jmp"   : true,
    "jmpq"  : true,
    "call"  : true,
    "callq" : true,
}

var _SegPrefix = map[string]_PrefixImpl {
    "cs": PrefixSegmentCS,
    "ds": PrefixSegmentDS,
    "es": PrefixSegmentES,
    "fs": PrefixSegmentFS,
    "gs": PrefixSegmentGS,
    "ss": PrefixSegmentSS,
}

type _ParserImpl struct {
    lex *asm.Tokenizer
}

func mkparser(lex *asm.Tokenizer) *_ParserImpl {
    return &_ParserImpl { lex }
}

func (self *_ParserImpl) i32(tk asm.Token, v int64) int32 {
    if v >= math.MinInt32 && v <= math.MaxUint32 {
        return int32(v)
    } else {
        panic(self.err(tk.Pos, fmt.Sprintf("32-bit integer out ouf range: %d", v)))
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

func (self *_ParserImpl) negv() int64 {
    tk := self.lex.Read()
    tt := tk.Ty

    /* must be an integer */
    if tt != asm.TokenInt {
        panic(self.err(tk.Pos, "integer expected after '-'"))
    } else {
        return -int64(tk.Uint)
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
            case '(' : q++; n++
            case ')' : q++; n--
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

    /* skip the last ")" */
    self.lex.Pos = q
    return r
}

func (self *_ParserImpl) relx(tk asm.Token) {
    if tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncLBrace {
        panic(self.err(tk.Pos, "'(' expected for RIP-relative addressing"))
    } else if tk = self.lex.Next(); self.regx(tk) != rip {
        panic(self.err(tk.Pos, "RIP-relative addressing expects %rip as the base register"))
    } else if tk = self.lex.Next(); tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncRBrace {
        panic(self.err(tk.Pos, "RIP-relative addressing does not support indexing or scaling"))
    }
}

func (self *_ParserImpl) immx(tk asm.Token) int64 {
    if tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncDollar {
        panic(self.err(tk.Pos, "'$' expected for registers"))
    } else if tk = self.lex.Read(); tk.Ty == asm.TokenInt {
        return int64(tk.Uint)
    } else if tk.Ty == asm.TokenPunc && tk.Punc() == asm.PuncLBrace {
        return self.eval(self.lex.Pos)
    } else if tk.Ty == asm.TokenPunc && tk.Punc() == asm.PuncMinus {
        return self.negv()
    } else {
        panic(self.err(tk.Pos, "immediate value expected"))
    }
}

func (self *_ParserImpl) regx(tk asm.Token) asm.Register {
    if tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncPercent {
        panic(self.err(tk.Pos, "'%' expected for registers"))
    } else if tk = self.lex.Read(); tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "register name expected"))
    } else if tk.Str == "rip" {
        return rip
    } else if reg, ok := Registers[tk.Str]; ok {
        return reg
    } else {
        panic(self.err(tk.Pos, "invalid register name: " + strconv.Quote(tk.Str)))
    }
}

func (self *_ParserImpl) regv(tk asm.Token) asm.Register {
    if reg := self.regx(tk); reg == rip {
        panic(self.err(tk.Pos, "%rip is not accessable as a dedicated register"))
    } else {
        return reg
    }
}

func (self *_ParserImpl) disp(vv int32) asm.MemoryAddress {
    switch tk := self.lex.Next(); tk.Ty {
        case asm.TokenEnd  : return MemoryAddress(nil, nil, 0, vv)
        case asm.TokenPunc : return self.relm(tk, vv)
        default            : panic(self.err(tk.Pos, "',' or '(' expected"))
    }
}

func (self *_ParserImpl) relm(tv asm.Token, disp int32) asm.MemoryAddress {
    var tk asm.Token
    var tt asm.TokenKind

    /* check for absolute addressing */
    if tv.Punc() == asm.PuncComma {
        self.lex.Pos--
        return MemoryAddress(nil, nil, 0, disp)
    }

    /* must be "(" now */
    if tv.Punc() != asm.PuncLBrace {
        panic(self.err(tv.Pos, "',' or '(' expected"))
    }

    /* read the next token */
    tk = self.lex.Next()
    tt = tk.Ty

    /* must be a punctuation */
    if tt != asm.TokenPunc {
        panic(self.err(tk.Pos, "'%' or ',' expected"))
    }

    /* check for base */
    switch tk.Punc() {
        case asm.PuncPercent : return self.base(tk, disp)
        case asm.PuncComma   : return self.index(nil, disp)
        default              : panic(self.err(tk.Pos, "'%' or ',' expected"))
    }
}

func (self *_ParserImpl) base(tk asm.Token, disp int32) asm.MemoryAddress {
    rr := self.regx(tk)
    nk := self.lex.Next()

    /* check for register indirection or base-index addressing */
    if !isReg64(rr) {
        panic(self.err(tk.Pos, "not a valid base register"))
    } else if nk.Ty != asm.TokenPunc {
        panic(self.err(nk.Pos, "',' or ')' expected"))
    } else if nk.Punc() == asm.PuncComma {
        return self.index(rr, disp)
    } else if nk.Punc() == asm.PuncRBrace {
        return MemoryAddress(rr, nil, 0, disp)
    } else {
        panic(self.err(nk.Pos, "',' or ')' expected"))
    }
}

func (self *_ParserImpl) index(base asm.Register, disp int32) asm.MemoryAddress {
    tk := self.lex.Next()
    rr := self.regx(tk)
    nk := self.lex.Next()

    /* check for scaled indexing */
    if base == rip {
        panic(self.err(tk.Pos, "RIP-relative addressing does not support indexing or scaling"))
    } else if !isIndexable(rr) {
        panic(self.err(tk.Pos, "not a valid index register"))
    } else if nk.Ty != asm.TokenPunc {
        panic(self.err(nk.Pos, "',' or ')' expected"))
    } else if nk.Punc() == asm.PuncComma {
        return self.scale(base, rr, disp)
    } else if nk.Punc() == asm.PuncRBrace {
        return MemoryAddress(base, rr, 1, disp)
    } else {
        panic(self.err(nk.Pos, "',' or ')' expected"))
    }
}

func (self *_ParserImpl) scale(base asm.Register, index asm.Register, disp int32) asm.MemoryAddress {
    tk := self.lex.Next()
    tt := tk.Ty
    tv := tk.Uint

    /* must be an integer */
    if tt != asm.TokenInt {
        panic(self.err(tk.Pos, "integer expected"))
    }

    /* scale can only be 1, 2, 4 or 8 */
    if tv == 0 || (_Scales & (1 << tv)) == 0 {
        panic(self.err(tk.Pos, "scale can only be 1, 2, 4 or 8"))
    }

    /* read next token */
    tk = self.lex.Next()
    tt = tk.Ty

    /* check for the closing ")" */
    if tt != asm.TokenPunc || tk.Punc() != asm.PuncRBrace {
        panic(self.err(tk.Pos, "')' expected"))
    }

    /* construct the memory address */
    return MemoryAddress(
        base,
        index,
        uint8(tv),
        disp,
    )
}

func (self *_ParserImpl) parse(ins *asm.ParsedInstruction) {
    var rr bool
    var lk bool
    var tt asm.TokenKind

    /* parse the first token */
    ff := true
    tk := self.lex.Next()

    /* special case for the "lock" prefix */
    if tk.Ty == asm.TokenName && strings.ToLower(tk.Str) == "lock" {
        lk = true
        tk = self.lex.Next()
    }

    /* must be an instruction */
    if tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "identifier expected"))
    }

    /* set the mnemonic, and check for LOCK prefix */
    if ins.Mnemonic = strings.ToLower(tk.Str); lk {
        ins.Prefixes = append(ins.Prefixes, PrefixLock)
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

        /* encountered an integer, must be a SIB memory address */
        if tt == asm.TokenInt {
            ins.Mem(self.disp(self.i32(tk, int64(tk.Uint))))
            continue
        }

        /* encountered an identifier, maybe an expression or a jump target, or a segment override prefix */
        if tt == asm.TokenName {
            ts := tk.Str
            tp := self.lex.Pos

            /* if the next token is EOF or a comma, it's a jumpt target */
            if tk = self.lex.Next(); tk.Ty == asm.TokenEnd || (tk.Ty == asm.TokenPunc && tk.Punc() == asm.PuncComma) {
                self.lex.Pos = tp
                ins.Target(ts)
                continue
            }

            /* if it is a colon, it's a segment override prefix, otherwise it must be an RIP-relative addressing operand */
            if tk.Ty != asm.TokenPunc || tk.Punc() != asm.PuncColon {
                self.relx(tk)
                ins.Reference(ts)
                continue
            }

            /* lookup segment prefixes */
            if p, ok := _SegPrefix[strings.ToLower(ts)]; !ok {
                panic(self.err(tk.Pos, "invalid segment name"))
            } else {
                ins.Prefixes = append(ins.Prefixes, p)
            }

            /* read the next token */
            tk = self.lex.Next()
            tt = tk.Ty

            /* encountered an integer, must be a SIB memory address */
            if tt == asm.TokenInt {
                ins.Mem(self.disp(self.i32(tk, int64(tk.Uint))))
                continue
            }
        }

        /* certain instructions may have a "*" before operands */
        if tt == asm.TokenPunc && tk.Punc() == asm.PuncStar {
            tk = self.lex.Next()
            tt = tk.Ty
            rr = true
        }

        /* ... otherwise it must be a punctuation */
        if tt != asm.TokenPunc {
            panic(self.err(tk.Pos, "'$', '%', '-' or '(' expected"))
        }

        /* check the operator */
        switch tk.Punc() {
            case asm.PuncLBrace  : break
            case asm.PuncMinus   : ins.Mem(self.disp(self.i32(tk, self.negv()))) ; continue
            case asm.PuncDollar  : ins.Imm(self.immx(tk))                        ; continue
            case asm.PuncPercent : ins.Reg(self.regv(tk))                        ; continue
            default              : panic(self.err(tk.Pos, "'$', '%', '-' or '(' expected"))
        }

        /* special case of '(', might be either `(expr)(SIB)` or just `(SIB)`
         * read one more token to confirm */
        tk = self.lex.Next()
        tt = tk.Ty

        /* the next token is '%', it's a memory address,
         * or ',' if it's a memory address without base,
         * otherwise it must be in `(expr)(SIB)` form */
        if tk.Ty == asm.TokenPunc && tk.Punc() == asm.PuncPercent {
            ins.Mem(self.base(tk, 0))
        } else if tk.Ty == asm.TokenPunc && tk.Punc() == asm.PuncComma {
            ins.Mem(self.index(nil, 0))
        } else {
            ins.Mem(self.disp(self.i32(tk, self.eval(tk.Pos))))
        }
    }

    /* check "jmp" and "call" instructions */
    if _RegBranch[ins.Mnemonic] {
        if len(ins.Operands) != 1 {
            panic(self.err(tk.Pos, fmt.Sprintf(`"%s" requires exact 1 argument`, ins.Mnemonic)))
        } else if !rr && ins.Operands[0].Op != asm.OpReg && ins.Operands[0].Op != asm.OpLabel {
            panic(self.err(tk.Pos, fmt.Sprintf(`invalid operand for "%s" instruction`, ins.Mnemonic)))
        }
    }
}
