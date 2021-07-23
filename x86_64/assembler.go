package x86_64

import (
    `fmt`
    `math`
    `strconv`
    `strings`
    `unicode`

    `github.com/chenzhuoyu/iasm/expr`
)

type (
	_TokenKind   int
    _Punctuation int
)

const (
    _T_end _TokenKind = iota + 1
    _T_int
    _T_name
    _T_punc
    _T_space
)

const (
    _P_plus _Punctuation = iota + 1
    _P_minus
    _P_star
    _P_slash
    _P_percent
    _P_amp
    _P_bar
    _P_caret
    _P_shl
    _P_shr
    _P_tilde
    _P_lbrk
    _P_rbrk
    _P_comma
    _P_dollar
)

var _PUNC_NAME = map[_Punctuation]string {
    _P_plus    : "+",
    _P_minus   : "-",
    _P_star    : "*",
    _P_slash   : "/",
    _P_percent : "%",
    _P_amp     : "&",
    _P_bar     : "|",
    _P_caret   : "^",
    _P_shl     : "<<",
    _P_shr     : ">>",
    _P_tilde   : "~",
    _P_lbrk    : "(",
    _P_rbrk    : ")",
    _P_comma   : ",",
    _P_dollar  : "$",
}

func (self _Punctuation) String() string {
    if v, ok := _PUNC_NAME[self]; ok {
        return v
    } else {
        return fmt.Sprintf("_Punctuation(%d)", self)
    }
}

type _Token struct {
    pos int
    end int
    u64 uint64
    str string
    tag _TokenKind
}

func (self *_Token) punc() _Punctuation {
    return _Punctuation(self.u64)
}

func (self *_Token) String() string {
    switch self.tag {
        case _T_end   : return "<END>"
        case _T_int   : return fmt.Sprintf("<INT %d>", self.u64)
        case _T_punc  : return fmt.Sprintf("<PUNC %s>", _Punctuation(self.u64))
        case _T_name  : return fmt.Sprintf("<NAME %s>", strconv.QuoteToASCII(self.str))
        case _T_space : return "<SPACE>"
        default       : return fmt.Sprintf("<UNK:%d %d %s>", self.tag, self.u64, strconv.QuoteToASCII(self.str))
    }
}

func tokenEnd(p int, end int) _Token {
    return _Token {
        pos: p,
        end: end,
        tag: _T_end,
    }
}

func tokenInt(p int, val uint64) _Token {
    return _Token {
        pos: p,
        u64: val,
        tag: _T_int,
    }
}

func tokenName(p int, name string) _Token {
    return _Token {
        pos: p,
        str: name,
        tag: _T_name,
    }
}

func tokenPunc(p int, punc _Punctuation) _Token {
    return _Token {
        pos: p,
        tag: _T_punc,
        u64: uint64(punc),
    }
}

func tokenSpace(p int, end int) _Token {
    return _Token {
        pos: p,
        end: end,
        tag: _T_space,
    }
}

// SyntaxError represents an error in the assembly syntax.
type SyntaxError struct {
    Pos    int
    Row    int
    Src    string
    Reason string
}

// Error implements the error interface.
func (self *SyntaxError) Error() string {
    return fmt.Sprintf("%s at %d:%d", self.Reason, self.Row, self.Pos + 1)
}

type _Tokenizer struct {
    pos int
    row int
    src []rune
}

func (self *_Tokenizer) ch() rune {
    return self.src[self.pos]
}

func (self *_Tokenizer) eof() bool {
    return self.pos >= len(self.src)
}

func (self *_Tokenizer) rch() (ret rune) {
    ret, self.pos = self.src[self.pos], self.pos + 1
    return
}

func (self *_Tokenizer) err(pos int, msg string) *SyntaxError {
    return &SyntaxError {
        Pos    : pos,
        Row    : self.row,
        Src    : string(self.src),
        Reason : msg,
    }
}

func (self *_Tokenizer) skip(check func(v rune) bool) {
    for !self.eof() && check(self.ch()) {
        self.pos++
    }
}

func (self *_Tokenizer) find(pos int, check func(v rune) bool) string {
    self.skip(check)
    return string(self.src[pos:self.pos])
}

func (self *_Tokenizer) chrv(p int) _Token {
    var err error
    var val uint64

    /* starting and ending position */
    p0 := p + 1
    p1 := p0 + 1

    /* find the end of the literal */
    for p1 < len(self.src) && self.src[p1] != '\'' {
        if p1++; self.src[p1 - 1] == '\\' {
            p1++
        }
    }

    /* empty literal */
    if p1 == p0 {
        panic(self.err(p1, "empty character constant"))
    }

    /* check for EOF */
    if p1 == len(self.src) {
        panic(self.err(p1, "unexpected EOF when scanning literals"))
    }

    /* parse the literal */
    if val, err = literal64(string(self.src[p0:p1])); err != nil {
        panic(self.err(p0, "cannot parse literal: " + err.Error()))
    }

    /* skip the closing '\'' */
    self.pos = p1 + 1
    return tokenInt(p, val)
}

func (self *_Tokenizer) numv(p int) _Token {
    if val, err := strconv.ParseUint(self.find(p, isnumber), 0, 64); err != nil {
        panic(self.err(p, "invalid immediate value: " + err.Error()))
    } else {
        return tokenInt(p, val)
    }
}

func (self *_Tokenizer) defv(p int, cc rune) _Token {
    if isdigit(cc) {
        return self.numv(p)
    } else if isident0(cc) {
        return tokenName(p, self.find(p, isident))
    } else {
        panic(self.err(p, "invalid char: " + strconv.QuoteRune(cc)))
    }
}

func (self *_Tokenizer) rep2(p int, pp _Punctuation, cc rune) _Token {
    if self.eof() {
        panic(self.err(self.pos, "unexpected EOF when scanning operators"))
    } else if c := self.rch(); c != cc {
        panic(self.err(p + 1, strconv.QuoteRune(cc) + " expected, got " + strconv.QuoteRune(c)))
    } else {
        return tokenPunc(p, pp)
    }
}

func (self *_Tokenizer) read() _Token {
    var p int
    var c rune
    var t _Token

    /* check for EOF */
    if self.eof() {
        return tokenEnd(self.pos, self.pos)
    }

    /* skip spaces as needed */
    if p = self.pos; unicode.IsSpace(self.src[p]) {
        self.skip(unicode.IsSpace)
        return tokenSpace(p, self.pos)
    }

    /* check for line comments */
    if p = self.pos; p < len(self.src) - 1 && self.src[p] == '/' && self.src[p + 1] == '/' {
        self.pos = len(self.src)
        return tokenEnd(p, self.pos)
    }

    /* read the next character */
    p = self.pos
    c = self.rch()

    /* parse the next character */
    switch c {
        case '+'  : t = tokenPunc(p, _P_plus)
        case '-'  : t = tokenPunc(p, _P_minus)
        case '*'  : t = tokenPunc(p, _P_star)
        case '/'  : t = tokenPunc(p, _P_slash)
        case '%'  : t = tokenPunc(p, _P_percent)
        case '&'  : t = tokenPunc(p, _P_amp)
        case '|'  : t = tokenPunc(p, _P_bar)
        case '^'  : t = tokenPunc(p, _P_caret)
        case '<'  : t = self.rep2(p, _P_shl, '<')
        case '>'  : t = self.rep2(p, _P_shr, '>')
        case '~'  : t = tokenPunc(p, _P_tilde)
        case '('  : t = tokenPunc(p, _P_lbrk)
        case ')'  : t = tokenPunc(p, _P_rbrk)
        case ','  : t = tokenPunc(p, _P_comma)
        case '$'  : t = tokenPunc(p, _P_dollar)
        case '\'' : t = self.chrv(p)
        default   : t = self.defv(p, c)
    }

    /* mark the end of token */
    t.end = self.pos
    return t
}

func (self *_Tokenizer) next() (tk _Token) {
    for {
        if tk = self.read(); tk.tag != _T_space {
            return
        }
    }
}

// LabelKind indicates the type of a label reference.
type LabelKind int

// OperandKind indicates the type of the operand.
type OperandKind int

const (
    // OpReg means the operand is a register.
    OpReg OperandKind = 1 << iota

    // OpImm means the operand is an immediate value.
    OpImm

    // OpMem means the operand is a memory address.
    OpMem

    // OpLabel means the operand is a label, specifically for
    // branch instructions.
    OpLabel
)

const (
    // JumpTarget means the label should be treated as a branch target.
    JumpTarget LabelKind = iota + 1

    // CodeReference means the label should be treated as a reference to
    // the code section (e.g. RIP-relative addressing).
    CodeReference
)

// ParsedLabel represents a label in the source, either a jump target or
// an RIP-relative addressing.
type ParsedLabel struct {
    Name string
    Kind LabelKind
}

// ParsedOperand represents an operand of an instruction in the source.
type ParsedOperand struct {
    Kind   OperandKind
    Imm    int64
    Reg    Register
    Label  ParsedLabel
    Memory MemoryAddress
}

// ParsedInstruction represents an instruction in the source.
type ParsedInstruction struct {
    Mnemonic string
    Operands []ParsedOperand
}

func (self *ParsedInstruction) imm(v int64) {
    self.Operands = append(self.Operands, ParsedOperand {
        Imm  : v,
        Kind : OpImm,
    })
}

func (self *ParsedInstruction) reg(v Register) {
    self.Operands = append(self.Operands, ParsedOperand {
        Reg  : v,
        Kind : OpReg,
    })
}

func (self *ParsedInstruction) mem(v MemoryAddress) {
    self.Operands = append(self.Operands, ParsedOperand {
        Kind   : OpMem,
        Memory : v,
    })
}

func (self *ParsedInstruction) target(v string) {
    self.Operands = append(self.Operands, ParsedOperand {
        Kind  : OpLabel,
        Label : ParsedLabel {
            Name: v,
            Kind: JumpTarget,
        },
    })
}

func (self *ParsedInstruction) reference(v string) {
    self.Operands = append(self.Operands, ParsedOperand {
        Kind  : OpLabel,
        Label : ParsedLabel {
            Name: v,
            Kind: CodeReference,
        },
    })
}

// LineKind indicates the type of a ParsedLine.
type LineKind int

// CommandArgKind indicates the type of a ParsedCommandArg.
type CommandArgKind int

const (
    // LineInstr means the ParsedLine is an instruction.
    LineInstr LineKind = iota + 1

    // LineLabel means the ParsedLine is a label.
    LineLabel

    // LineCommand means the ParsedLine is a ParsedCommand.
    LineCommand
)

const (
    // ArgInt means the ParsedCommandArg is an integer.
    ArgInt CommandArgKind = iota + 1

    // ArgString means the ParsedCommandArg is a string.
    ArgString

    // ArgExpression means the ParsedCommandArg is an expression.
    ArgExpression
)

// ParsedLine represents a parsed source line.
type ParsedLine struct {
    Kind        LineKind
    Label       string
    Command     ParsedCommand
    Instruction ParsedInstruction
}

// ParsedCommand represents a parsed assembly directive command.
type ParsedCommand struct {
    Cmd  string
    Args []ParsedCommandArg
}

// ParsedCommandArg represents an argument of a ParsedCommand.
type ParsedCommandArg struct {
    Int  int64
    Str  string
    Kind CommandArgKind
}

// Parser parses the source, and generates a sequence of ParsedInstruction's.
type Parser struct {
    row int
    src string
    lex _Tokenizer
    exp expr.Expression
}

const (
    rip Register64 = 0xff
)

func (self *Parser) i32(tk _Token, v int64) int32 {
    if v >= math.MinInt32 && v <= math.MaxUint32 {
        return int32(v)
    } else {
        panic(self.err(tk.pos, fmt.Sprintf("32-bit integer out ouf range: %d", v)))
    }
}

func (self *Parser) err(pos int, msg string) *SyntaxError {
    return &SyntaxError {
        Pos    : pos,
        Row    : self.row,
        Src    : self.src,
        Reason : msg,
    }
}

func (self *Parser) negv() int64 {
    tk := self.lex.read()
    tt := tk.tag

    /* must be an integer */
    if tt != _T_int {
        panic(self.err(tk.pos, "integer expected after '-'"))
    } else {
        return -int64(tk.u64)
    }
}

func (self *Parser) eval(p int) int64 {
    var v int64
    var e error

    /* searching start */
    n := 1
    q := p + 1

    /* find the end of expression */
    for n > 0 && q < len(self.lex.src) {
        switch self.lex.src[q] {
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
    if v, e = self.exp.SetSource(string(self.lex.src[p:q - 1])).Evaluate(nil); e != nil {
        panic(self.err(p, "cannot evaluate expression: " + e.Error()))
    }

    /* skip the last ')' */
    self.lex.pos = q
    return v
}

func (self *Parser) relx(tk _Token) {
    if tk.tag != _T_punc || tk.punc() != _P_lbrk {
        panic(self.err(tk.pos, "'(' expected for RIP-relative addressing"))
    } else if tk = self.lex.next(); self.regx(tk) != rip {
        panic(self.err(tk.pos, "RIP-relative addressing expects %rip as the base register"))
    } else if tk = self.lex.next(); tk.tag != _T_punc || tk.punc() != _P_rbrk {
        panic(self.err(tk.pos, "RIP-relative addressing does not support indexing or scaling"))
    }
}

func (self *Parser) immx(tk _Token) int64 {
    if tk.tag != _T_punc || tk.punc() != _P_dollar {
        panic(self.err(tk.pos, "'$' expected for registers"))
    } else if tk = self.lex.read(); tk.tag == _T_int {
        return int64(tk.u64)
    } else if tk.tag == _T_punc && tk.punc() == _P_lbrk {
        return self.eval(self.lex.pos)
    } else if tk.tag == _T_punc && tk.punc() == _P_minus {
        return self.negv()
    } else {
        panic(self.err(tk.pos, "immediate value expected"))
    }
}

func (self *Parser) regx(tk _Token) Register {
    if tk.tag != _T_punc || tk.punc() != _P_percent {
        panic(self.err(tk.pos, "'%' expected for registers"))
    } else if tk = self.lex.read(); tk.tag != _T_name {
        panic(self.err(tk.pos, "register name expected"))
    } else if tk.str == "rip" {
        return rip
    } else if reg, ok := Registers[tk.str]; ok {
        return reg
    } else {
        panic(self.err(tk.pos, "invalid register name: " + tk.str))
    }
}

func (self *Parser) regv(tk _Token) Register {
    if reg := self.regx(tk); reg == rip {
        panic(self.err(tk.pos, "%rip is not accessable as a dedicated register"))
    } else {
        return reg
    }
}

func (self *Parser) disp(vv int32) MemoryAddress {
    tk := self.lex.next()
    tt := tk.tag

    /* must be an opening '(' */
    if tt != _T_punc || tk.punc() != _P_lbrk {
        panic(self.err(tk.pos, "'(' expected"))
    }

    /* read the next token */
    tk = self.lex.next()
    tt = tk.tag

    /* must be a punctuation */
    if tt != _T_punc {
        panic(self.err(tk.pos, "'%' or ',' expected"))
    }

    /* check for base */
    switch tk.punc() {
        case _P_percent : return self.base(tk, vv)
        case _P_comma   : return self.index(nil, vv)
        default         : panic(self.err(tk.pos, "'%' or ',' expected"))
    }
}

func (self *Parser) base(tk _Token, disp int32) MemoryAddress {
    rr := self.regx(tk)
    nk := self.lex.next()

    /* check for register indirection or base-index addressing */
    if nk.tag != _T_punc {
        panic(self.err(nk.pos, "',' or ')' expected"))
    } else if nk.punc() == _P_comma {
        return self.index(rr, disp)
    } else if nk.punc() == _P_rbrk {
        return MemoryAddress{Base: rr, Displacement: disp}
    } else {
        panic(self.err(nk.pos, "',' or ')' expected"))
    }
}

func (self *Parser) index(base Register, disp int32) MemoryAddress {
    rr := self.regx(self.lex.next())
    tk := self.lex.next()

    /* check for scaled indexing */
    if tk.tag != _T_punc {
        panic(self.err(tk.pos, "',' or ')' expected"))
    } else if tk.punc() == _P_comma {
        return self.scale(base, rr, disp)
    } else if tk.punc() == _P_rbrk {
        return MemoryAddress{Base: base, Index: rr, Scale: 1, Displacement: disp}
    } else {
        panic(self.err(tk.pos, "',' or ')' expected"))
    }
}

func (self *Parser) scale(base Register, index Register, disp int32) MemoryAddress {
    tk := self.lex.next()
    tt := tk.tag
    tv := tk.u64

    /* must be an integer */
    if tt != _T_int {
        panic(self.err(tk.pos, "integer expected"))
    }

    /* scale can only be 1, 2, 4 or 8 */
    if tv == 0 || (_Scales & (1 << tv)) == 0 {
        panic(self.err(tk.pos, "scale can only be 1, 2, 4 or 8"))
    }

    /* read next token */
    tk = self.lex.next()
    tt = tk.tag

    /* check for the closing ')' */
    if tt != _T_punc || tk.punc() != _P_rbrk {
        panic(self.err(tk.pos, "')' expected"))
    }

    /* construct the memory address */
    return MemoryAddress {
        Base         : base,
        Index        : index,
        Scale        : uint8(tv),
        Displacement : disp,
    }
}

func (self *Parser) feed(line string) *ParsedLine {
    self.row++
    self.src = line

    /* check for directives */
    if line[0] == '.' {
        return self.cmds(line)
    }

    /* check for labels */
    if line[len(line) - 1] == ':' {
        return self.labels(line)
    }

    /* it must be instructions now */
    self.lex.pos = 0
    self.lex.row = self.row
    self.lex.src = []rune(line)

    /* parse the first token */
    tk := self.lex.next()
    tt := tk.tag
    ff := true

    /* the first token must be the mnemonic */
    if tt != _T_name {
        panic(self.err(tk.pos, "mnemonic expected"))
    }

    /* set the line kind and mnemonic */
    ret := &ParsedLine {
        Kind        : LineInstr,
        Instruction : ParsedInstruction{Mnemonic: tk.str},
    }

    /* parse all the operands */
    for {
        tk = self.lex.next()
        tt = tk.tag

        /* check for end of line */
        if tt == _T_end {
            return ret
        }

        /* expect a comma if not the first operand */
        if !ff {
            if tt == _T_punc && tk.punc() == _P_comma {
                tk = self.lex.next()
            } else {
                panic(self.err(tk.pos, "',' expected"))
            }
        }

        /* not the first operand anymore */
        ff = false
        tt = tk.tag

        /* encountered an integer, must be a SIB memory address */
        if tt == _T_int {
            ret.Instruction.mem(self.disp(self.i32(tk, int64(tk.u64))))
            continue
        }

        /* encountered an identifier, maybe an expression or a jump target */
        if tt == _T_name {
            ts := tk.str
            tp := self.lex.pos

            /* if the next token is EOF or a comma, it's a jumpt target */
            if tk = self.lex.next(); tk.tag == _T_end || (tk.tag == _T_punc && tk.punc() == _P_comma) {
                self.lex.pos = tp
                ret.Instruction.target(ts)
                continue
            }

            /* otherwise it must be an RIP-relative addressing operand */
            self.relx(tk)
            ret.Instruction.reference(ts)
            continue
        }

        /* otherwise it must be a punctuation */
        if tt != _T_punc {
            panic(self.err(tk.pos, "'$', '%', '-' or '(' expected"))
        }

        /* check the operator */
        switch tk.punc() {
            case _P_lbrk    : break
            case _P_minus   : ret.Instruction.mem(self.disp(self.i32(tk, self.negv()))) ; continue
            case _P_dollar  : ret.Instruction.imm(self.immx(tk))                        ; continue
            case _P_percent : ret.Instruction.reg(self.regv(tk))                        ; continue
            default         : panic(self.err(tk.pos, "'$', '%', '-' or '(' expected"))
        }

        /* special case of '(', might be either `(expr)(SIB)` or just `(SIB)`
         * read one more token to confirm */
        tk = self.lex.next()
        tt = tk.tag

        /* the next token is '%', it's a memory address,
         * or ',' if it's a memory address without base,
         * otherwise it must be in `(expr)(SIB)` form */
        if tk.tag == _T_punc && tk.punc() == _P_percent {
            ret.Instruction.mem(self.base(tk, 0))
        } else if tk.tag == _T_punc && tk.punc() == _P_comma {
            ret.Instruction.mem(self.index(nil, 0))
        } else {
            ret.Instruction.mem(self.disp(self.i32(tk, self.eval(tk.pos))))
        }
    }
}

func (self *Parser) cmds(line string) *ParsedLine {
    var p int
    var v []ParsedCommandArg

    /* split the commands */
    vals := strings.SplitN(line, " ", 2)
    args := []rune(strings.Join(vals[1:], " "))

    /* parse the arguments */
    loop: for {
        switch self.next(args, &p) {
            case 0   : break loop
            case '"' : p = self.strings(&v, args, p)
            case '-' : fallthrough
            case '0' : fallthrough
            case '1' : fallthrough
            case '2' : fallthrough
            case '3' : fallthrough
            case '4' : fallthrough
            case '5' : fallthrough
            case '6' : fallthrough
            case '7' : fallthrough
            case '8' : fallthrough
            case '9' : p = self.integers(&v, args, p)
            default  : p = self.expressions(&v, args, p)
        }
    }

    /* construct the line */
    return &ParsedLine {
        Kind    : LineCommand,
        Command : ParsedCommand {
            Cmd  : vals[0],
            Args : v,
        },
    }
}

func (self *Parser) next(line []rune, p *int) rune {
    for {
        if *p >= len(line) {
            return 0
        } else if cc := line[*p]; !unicode.IsSpace(cc) {
            return cc
        } else {
            *p++
        }
    }
}

func (self *Parser) labels(line string) *ParsedLine {
    return &ParsedLine {
        Kind  : LineLabel,
        Label : line[:len(line) - 1],
    }
}

func (self *Parser) strings(argv *[]ParsedCommandArg, line []rune, p int) int {
    var i int
    var e error
    var v string

    /* find the end of string */
    for i = p + 1; i < len(line) && line[i] != '"'; i++ {
        if line[i] == '\\' {
            i++
        }
    }

    /* check for EOF */
    if i == len(line) {
        panic(self.err(i, "unexpected EOF when scanning strings"))
    }

    /* unquote the string */
    if v, e = strconv.Unquote(string(line[p:i + 1])); e != nil {
        panic(self.err(p, "invalid string: " + e.Error()))
    }

    /* add the argument to buffer */
    *argv = append(*argv, ParsedCommandArg{Str: v, Kind: ArgString})
    return i + 1
}

func (self *Parser) integers(argv *[]ParsedCommandArg, line []rune, p int) int {
    var i int
    var v int64
    var e error

    /* skip the '-' if needed */
    if line[p] != '-' {
        i = p
    } else {
        i = p + 1
    }

    /* find the end of integer */
    for i < len(line) && isnumber(line[i]) {
        i++
    }

    /* parse the integer */
    if v, e = strconv.ParseInt(string(line[p:i]), 0, 64); e != nil {
        panic(self.err(p, "invalid integer: " + e.Error()))
    }

    /* add the argument to buffer */
    *argv = append(*argv, ParsedCommandArg{Int: v, Kind: ArgInt})
    return i
}

func (self *Parser) expressions(argv *[]ParsedCommandArg, line []rune, p int) int {
    var i int
    var n int
    var s int

    /* scan until the first standalone ',' or EOF */
    loop: for i = p; i < len(line); i++ {
        switch line[i] {
            case '"'           : s ^= 1
            case ','           : if s == 0 { if n == 0 { break loop } }
            case ']', '}', '>' : if s == 0 { if n == 0 { break loop } else { n-- } }
            case '[', '{', '<' : if s == 0 { n++ }
            case '\\'          : if s != 0 { i++ }
        }
    }

    /* check for EOF in strings */
    if s != 0 {
        panic(self.err(i, "unexpected EOF when scanning strings"))
    }

    /* check for bracket matching */
    if n != 0 {
        panic(self.err(i, "unbalanced '{' or '[' or '<'"))
    }

    /* add the argument to buffer */
    *argv = append(*argv, ParsedCommandArg{Str: string(line[p:i]), Kind: ArgExpression})
    return i
}

// Feed feeds the parser with one more line, and the parser
// parses it into a ParsedLine.
//
// NOTE: Feed does not handle empty lines or multiple lines,
//       it panics when this happens.
//
func (self *Parser) Feed(line string) *ParsedLine {
    if strings.TrimSpace(line) == "" {
        panic("empty line")
    } else if strings.ContainsRune(line, '\n') {
        panic("passing multiple lines to Feed()")
    } else {
        return self.feed(line)
    }
}
