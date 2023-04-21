package x86_64

import (
    `fmt`
    `math`
    `strconv`
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
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

func (self *_ParserImpl) relx(tk asm.Token) {
    if !tk.IsPunc(asm.PuncLBrace) {
        panic(self.err(tk.Pos, "'(' expected for RIP-relative addressing"))
    } else if tk = self.lex.Next(); self.regx(tk) != rip {
        panic(self.err(tk.Pos, "RIP-relative addressing expects %rip as the base register"))
    } else if tk = self.lex.Next(); !tk.IsPunc(asm.PuncRBrace) {
        panic(self.err(tk.Pos, "RIP-relative addressing does not support indexing or scaling"))
    }
}

func (self *_ParserImpl) regx(tk asm.Token) asm.Register {
    if !tk.IsPunc(asm.PuncPercent) {
        panic(self.err(tk.Pos, "'%' expected for registers"))
    } else if tk = self.lex.Read(); tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "register name expected"))
    } else if key := strings.ToLower(tk.Str); key == "rip" {
        return rip
    } else if reg, ok := Registers[key]; ok {
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
    pos := self.lex.Pos
    row := self.lex.Row
    tok := self.lex.Next()

    /* absolute addressing, may require rewinding */
    if tok.Ty == asm.TokenEnd || tok.IsPunc(asm.PuncComma) {
        self.lex.Pos = pos
        self.lex.Row = row
        return MemoryAddress(nil, nil, 0, vv)
    }

    /* begin of memory address */
    if !tok.IsPunc(asm.PuncLBrace) {
        panic(self.err(tok.Pos, "'(' expected"))
    }

    /* memory address */
    if self.lex.MatchPunc(asm.PuncComma) {
        return self.index(nil, vv)
    } else {
        return self.base(self.lex.Next(), vv)
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
    bv := 1 << tk.Uint

    /* must be an integer */
    if tk.Ty != asm.TokenInt {
        panic(self.err(tk.Pos, "integer expected"))
    }

    /* check for scale value */
    if tk.Uint == 0 || _Scales & bv == 0 {
        panic(self.err(tk.Pos, "scale can only be 1, 2, 4 or 8"))
    }

    /* match the closing ")" */
    self.lex.ExpectPunc(asm.PuncRBrace)
    return MemoryAddress(base, index, uint8(tk.Uint), disp)
}

func (self *_ParserImpl) value(ins *asm.ParsedInstruction, rr *bool) {
    var ok bool
    var tk asm.Token
    var pp _PrefixImpl

    /* save the tokenizer state */
    pos := self.lex.Pos
    row := self.lex.Row

    /* encountered an integer, must be a SIB memory address */
    if tk = self.lex.Next(); tk.Ty == asm.TokenInt {
        ins.Mem(self.disp(self.i32(tk, int64(tk.Uint))))
        return
    }

    /* encountered an identifier, maybe an expression or a jump target, or a segment override prefix */
    if tk.Ty == asm.TokenName {
        ts := tk.Str
        tp := self.lex.Pos
        tr := self.lex.Row

        /* if the next token is EOF or a comma, it's a jumpt target */
        if tk = self.lex.Next(); tk.Ty == asm.TokenEnd || tk.IsPunc(asm.PuncComma) {
            self.lex.Pos = tp
            self.lex.Row = tr
            ins.Target(ts)
            return
        }

        /* if it is a colon, it's a segment override prefix, otherwise it must be an RIP-relative addressing operand */
        if !tk.IsPunc(asm.PuncColon) {
            self.relx(tk)
            ins.Reference(ts)
            return
        }

        /* lookup segment prefixes */
        if pp, ok = _SegPrefix[strings.ToLower(ts)]; !ok {
            panic(self.err(tk.Pos, "invalid segment name"))
        }

        /* read the next token */
        tk = self.lex.Next()
        ins.Prefixes = append(ins.Prefixes, pp)

        /* encountered an integer, must be a SIB memory address */
        if tk.Ty == asm.TokenInt {
            ins.Mem(self.disp(self.i32(tk, int64(tk.Uint))))
            return
        }
    }

    /* certain instructions may have a "*" before operands */
    if tk.IsPunc(asm.PuncStar) {
        tk = self.lex.Next()
        *rr = true
    }

    /* ... otherwise it must be a punctuation */
    if tk.Ty != asm.TokenPunc {
        panic(self.err(tk.Pos, "'$', '%', '-' or '(' expected"))
    }

    /* '(' is special because it might be either `(expr)(SIB)` or just `(SIB)`,
     * read one more token to confirm, handle it later */
    if tk.Punc() != asm.PuncLBrace {
        switch tk.Punc() {
            case asm.PuncMinus   : ins.Mem(self.disp(self.i32(tk, -self.lex.Value(nil)))) ; return
            case asm.PuncDollar  : ins.Imm(self.lex.Value(nil))                           ; return
            case asm.PuncPercent : ins.Reg(self.regv(tk))                                 ; return
            default              : panic(self.err(tk.Pos, "'$', '%', '-' or '(' expected"))
        }
    }

    /* the next token is '%', it's a memory address, or ',' if it's a memory address without base */
    if tk = self.lex.Next(); tk.Ty == asm.TokenPunc {
        switch tk.Punc() {
            case asm.PuncPercent : ins.Mem(self.base(tk, 0))   ; return
            case asm.PuncComma   : ins.Mem(self.index(nil, 0)) ; return
        }
    }

    /* ... otherwise it must be `(expr)(SIB)`, revert the tokenizer */
    self.lex.Pos = pos
    self.lex.Row = row
    ins.Mem(self.disp(self.i32(tk, self.lex.Value(nil))))
}

func (self *_ParserImpl) parse(ins *asm.ParsedInstruction) {
    rr := false
    tk := self.lex.Next()

    /* special case for the "lock" prefix */
    if tk.Ty == asm.TokenName && strings.ToLower(tk.Str) == "lock" {
        tk = self.lex.Next()
        ins.Prefixes = append(ins.Prefixes, PrefixLock)
    }

    /* must be an instruction */
    if tk.Ty != asm.TokenName {
        panic(self.err(tk.Pos, "identifier expected"))
    }

    /* set the mnemonic */
    pos := tk.Pos
    ins.Mnemonic = strings.ToLower(tk.Str)

    /* check if the instruction has operands */
    if self.lex.MatchEnd() {
        return
    }

    /* construct a placeholder token */
    tk.Ty = asm.TokenPunc
    tk.Uint = uint64(asm.PuncComma)

    /* parse all the operands */
    for tk.IsPunc(asm.PuncComma) {
        self.value(ins, &rr)
        tk = self.lex.Next()
    }

    /* should be the end of instruction */
    if tk.Ty != asm.TokenEnd {
        panic(self.err(tk.Pos, "garbage after instruction"))
    }

    /* check "jmp" and "call" instructions */
    if _RegBranch[ins.Mnemonic] {
        if len(ins.Operands) != 1 {
            panic(self.err(pos, fmt.Sprintf(`"%s" requires exact 1 argument`, ins.Mnemonic)))
        } else if !rr && ins.Operands[0].Op != asm.OpReg && ins.Operands[0].Op != asm.OpLabel {
            panic(self.err(pos, fmt.Sprintf(`invalid operand for "%s" instruction`, ins.Mnemonic)))
        }
    }
}
