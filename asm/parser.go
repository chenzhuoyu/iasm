package asm

import (
    `errors`
    `fmt`
    `strconv`
    `strings`
    `unicode`
)

// TokenKind represents the type of a token.
type TokenKind int

const (
    TokenEnd TokenKind = iota + 1
    TokenInt
    TokenName
    TokenPunc
    TokenSpace
)

// Punctuation represents all the characters an operator may have.
type Punctuation int

const (
    PuncPlus Punctuation = iota + 1
    PuncMinus
    PuncStar
    PuncSlash
    PuncPercent
    PuncAmp
    PuncBar
    PuncCaret
    PuncShl
    PuncShr
    PuncTilde
    PuncLBrk
    PuncRBrk
    PuncDot
    PuncComma
    PuncColon
    PuncDollar
    PuncHash
    PuncBang
)

var _PUNC_NAME = map[Punctuation]string {
    PuncPlus    : "+",
    PuncMinus   : "-",
    PuncStar    : "*",
    PuncSlash   : "/",
    PuncPercent : "%",
    PuncAmp     : "&",
    PuncBar     : "|",
    PuncCaret   : "^",
    PuncShl     : "<<",
    PuncShr     : ">>",
    PuncTilde   : "~",
    PuncLBrk    : "(",
    PuncRBrk    : ")",
    PuncDot     : ".",
    PuncComma   : ",",
    PuncColon   : ":",
    PuncDollar  : "$",
    PuncHash    : "#",
    PuncBang    : "!",
}

func (self Punctuation) String() string {
    if v, ok := _PUNC_NAME[self]; ok {
        return v
    } else {
        return fmt.Sprintf("Punctuation(%d)", self)
    }
}

type Token struct {
    Ty   TokenKind
    Pos  int
    End  int
    Str  string
    Uint uint64
}

// Punc converts the token into a Punctuation.
func (self *Token) Punc() Punctuation {
    if self.Ty != TokenPunc {
        panic("asm: token is not a punctuation")
    } else {
        return Punctuation(self.Uint)
    }
}

func (self *Token) String() string {
    switch self.Ty {
        case TokenEnd   : return "<END>"
        case TokenInt   : return fmt.Sprintf("<INT %d>", self.Uint)
        case TokenPunc  : return fmt.Sprintf("<PUNC %s>", Punctuation(self.Uint))
        case TokenName  : return fmt.Sprintf("<NAME %s>", strconv.QuoteToASCII(self.Str))
        case TokenSpace : return "<SPACE>"
        default         : return fmt.Sprintf("<UNK:%d %d %s>", self.Ty, self.Uint, strconv.QuoteToASCII(self.Str))
    }
}

func tokenEnd(p int, end int) Token {
    return Token {
        Ty  : TokenEnd,
        Pos : p,
        End : end,
    }
}

func tokenInt(p int, val uint64) Token {
    return Token {
        Ty   : TokenInt,
        Pos  : p,
        Uint : val,
    }
}

func tokenName(p int, name string) Token {
    return Token {
        Ty  : TokenName,
        Pos : p,
        Str : name,
    }
}

func tokenPunc(p int, punc Punctuation) Token {
    return Token {
        Ty   : TokenPunc,
        Pos  : p,
        Uint : uint64(punc),
    }
}

func tokenSpace(p int, end int) Token {
    return Token {
        Ty  : TokenSpace,
        Pos : p,
        End : end,
    }
}

// SyntaxError represents an error in the assembly syntax.
type SyntaxError struct {
    Pos    int
    Row    int
    Src    []rune
    Reason string
}

// Error implements the error interface.
func (self *SyntaxError) Error() string {
    if self.Pos < 0 {
        return fmt.Sprintf("%s at line %d", self.Reason, self.Row)
    } else {
        return fmt.Sprintf("%s at %d:%d", self.Reason, self.Row, self.Pos + 1)
    }
}

type (
	_TrimState int
)

const (
    _TS_normal _TrimState = iota
    _TS_slcomm
    _TS_hscomm
    _TS_string
    _TS_escape
    _TS_accept
    _TS_nolast
)

// Tokenizer splits the source into token stream. It's fully unicode compatible.
type Tokenizer struct {
    Pos int
    Row int
    Src []rune
}

func (self *Tokenizer) ch() rune {
    return self.Src[self.Pos]
}

func (self *Tokenizer) eof() bool {
    return self.Pos >= len(self.Src)
}

func (self *Tokenizer) rch() (ret rune) {
    ret, self.Pos = self.Src[self.Pos], self.Pos+ 1
    return
}

func (self *Tokenizer) err(pos int, msg string) *SyntaxError {
    return &SyntaxError {
        Pos    : pos,
        Row    : self.Row,
        Src    : self.Src,
        Reason : msg,
    }
}

func (self *Tokenizer) init(src string) {
    var i int
    var ch rune
    var st _TrimState

    /* set the source */
    self.Pos = 0
    self.Src = []rune(src)

    /* remove commends, including "//" and "##" */
    loop: for i, ch = range self.Src {
        switch {
            case st == _TS_normal && ch == '/'  : st = _TS_slcomm
            case st == _TS_normal && ch == '"'  : st = _TS_string
            case st == _TS_normal && ch == ';'  : st = _TS_accept; break loop
            case st == _TS_normal && ch == '#'  : st = _TS_hscomm
            case st == _TS_slcomm && ch == '/'  : st = _TS_nolast; break loop
            case st == _TS_slcomm               : st = _TS_normal
            case st == _TS_hscomm && ch == '#'  : st = _TS_nolast; break loop
            case st == _TS_hscomm               : st = _TS_normal
            case st == _TS_string && ch == '"'  : st = _TS_normal
            case st == _TS_string && ch == '\\' : st = _TS_escape
            case st == _TS_escape               : st = _TS_string
        }
    }

    /* check for errors */
    switch st {
        case _TS_accept: self.Src = self.Src[:i]
        case _TS_nolast: self.Src = self.Src[:i - 1]
        case _TS_string: panic(self.err(i, "string is not terminated"))
        case _TS_escape: panic(self.err(i, "escape sequence is not terminated"))
    }
}

func (self *Tokenizer) skip(check func(v rune) bool) {
    for !self.eof() && check(self.ch()) {
        self.Pos++
    }
}

func (self *Tokenizer) find(pos int, check func(v rune) bool) string {
    self.skip(check)
    return string(self.Src[pos:self.Pos])
}

func (self *Tokenizer) chrv(p int) Token {
    var err error
    var val uint64

    /* starting and ending position */
    p0 := p + 1
    p1 := p0 + 1

    /* find the end of the literal */
    for p1 < len(self.Src) && self.Src[p1] != '\'' {
        if p1++; self.Src[p1 - 1] == '\\' {
            p1++
        }
    }

    /* empty literal */
    if p1 == p0 {
        panic(self.err(p1, "empty character constant"))
    }

    /* check for EOF */
    if p1 == len(self.Src) {
        panic(self.err(p1, "unexpected EOF when scanning literals"))
    }

    /* parse the literal */
    if val, err = literal64(string(self.Src[p0:p1])); err != nil {
        panic(self.err(p0, "cannot parse literal: " + err.Error()))
    }

    /* skip the closing '\'' */
    self.Pos = p1 + 1
    return tokenInt(p, val)
}

func (self *Tokenizer) numv(p int) Token {
    if val, err := strconv.ParseUint(self.find(p, isnumber), 0, 64); err != nil {
        panic(self.err(p, "invalid immediate value: " + err.Error()))
    } else {
        return tokenInt(p, val)
    }
}

func (self *Tokenizer) defv(p int, cc rune) Token {
    if isdigit(cc) {
        return self.numv(p)
    } else if isident0(cc) {
        return tokenName(p, self.find(p, isident))
    } else {
        panic(self.err(p, "invalid char: " + strconv.QuoteRune(cc)))
    }
}

func (self *Tokenizer) rep2(p int, pp Punctuation, cc rune) Token {
    if self.eof() {
        panic(self.err(self.Pos, "unexpected EOF when scanning operators"))
    } else if c := self.rch(); c != cc {
        panic(self.err(p + 1, strconv.QuoteRune(cc) + " expected, got " + strconv.QuoteRune(c)))
    } else {
        return tokenPunc(p, pp)
    }
}

// Read advances the tokenizer, and return the next token.
func (self *Tokenizer) Read() Token {
    var p int
    var c rune
    var t Token

    /* check for EOF */
    if self.eof() {
        return tokenEnd(self.Pos, self.Pos)
    }

    /* skip spaces as needed */
    if p = self.Pos; unicode.IsSpace(self.Src[p]) {
        self.skip(unicode.IsSpace)
        return tokenSpace(p, self.Pos)
    }

    /* check for line comments */
    if p = self.Pos; p < len(self.Src) - 1 && self.Src[p] == '/' && self.Src[p + 1] == '/' {
        self.Pos = len(self.Src)
        return tokenEnd(p, self.Pos)
    }

    /* read the next character */
    p = self.Pos
    c = self.rch()

    /* parse the next character */
    switch c {
        case '+'  : t = tokenPunc(p, PuncPlus)
        case '-'  : t = tokenPunc(p, PuncMinus)
        case '*'  : t = tokenPunc(p, PuncStar)
        case '/'  : t = tokenPunc(p, PuncSlash)
        case '%'  : t = tokenPunc(p, PuncPercent)
        case '&'  : t = tokenPunc(p, PuncAmp)
        case '|'  : t = tokenPunc(p, PuncBar)
        case '^'  : t = tokenPunc(p, PuncCaret)
        case '<'  : t = self.rep2(p, PuncShl, '<')
        case '>'  : t = self.rep2(p, PuncShr, '>')
        case '~'  : t = tokenPunc(p, PuncTilde)
        case '('  : t = tokenPunc(p, PuncLBrk)
        case ')'  : t = tokenPunc(p, PuncRBrk)
        case '.'  : t = tokenPunc(p, PuncDot)
        case ','  : t = tokenPunc(p, PuncComma)
        case ':'  : t = tokenPunc(p, PuncColon)
        case '$'  : t = tokenPunc(p, PuncDollar)
        case '#'  : t = tokenPunc(p, PuncHash)
        case '!'  : t = tokenPunc(p, PuncBang)
        case '\'' : t = self.chrv(p)
        default   : t = self.defv(p, c)
    }

    /* mark the end of token */
    t.End = self.Pos
    return t
}

// Next is like Read, but ignores all whitespace tokens.
func (self *Tokenizer) Next() (tk Token) {
    for {
        if tk = self.Read(); tk.Ty != TokenSpace {
            return
        }
    }
}

// LabelKind indicates the type of label reference.
type LabelKind int

// OperandKind indicates the type of the operand.
type OperandKind int

const (
    // OpImm means the operand is an immediate value.
    OpImm OperandKind = 1 << iota

    // OpReg means the operand is a register.
    OpReg

    // OpMem means the operand is a memory address.
    OpMem

    // OpLabel means the operand is a label, specifically for
    // branch instructions.
    OpLabel
)

const (
    // Declaration means the label is a declaration.
    Declaration LabelKind = iota + 1

    // BranchTarget means the label should be treated as a branch target.
    BranchTarget

    // RelativeAddress means the label should be treated as a reference to
    // the code section (e.g. RIP-relative addressing).
    RelativeAddress
)

// InstructionPrefix indicates the prefix bytes prepended to the instruction.
type InstructionPrefix interface {
    EncodePrefix(p *Instruction)
}

// ParsedLabel represents a label in the source, either a jump target or
// an RIP-relative addressing.
type ParsedLabel struct {
    Name string
    Kind LabelKind
}

// ParsedOperand represents an operand of an instruction in the source.
type ParsedOperand struct {
    Op     OperandKind
    Imm    int64
    Reg    Register
    Label  ParsedLabel
    Memory MemoryAddress
}

// ParsedInstruction represents an instruction in the source.
type ParsedInstruction struct {
    Mnemonic string
    Operands []ParsedOperand
    Prefixes []InstructionPrefix
}

// Imm adds an immediate operand to this instruction.
func (self *ParsedInstruction) Imm(v int64) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op  : OpImm,
        Imm : v,
    })
}

// Reg adds a register operand to this instruction.
func (self *ParsedInstruction) Reg(v Register) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op  : OpReg,
        Reg : v,
    })
}

// Mem adds a memory operand to this instruction.
func (self *ParsedInstruction) Mem(v MemoryAddress) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op     : OpMem,
        Memory : v,
    })
}

// Target adds a branch target operand to this instruction.
func (self *ParsedInstruction) Target(v string) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op    : OpLabel,
        Label : ParsedLabel {
            Name: v,
            Kind: BranchTarget,
        },
    })
}

// Reference adds a label reference operand to this instruction.
func (self *ParsedInstruction) Reference(v string) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op    : OpLabel,
        Label : ParsedLabel {
            Name: v,
            Kind: RelativeAddress,
        },
    })
}

// LineKind indicates the type of ParsedLine.
type LineKind int

const (
    // LineLabel means the ParsedLine is a label.
    LineLabel LineKind = iota + 1

    // LineInstr means the ParsedLine is an instruction.
    LineInstr

    // LineCommand means the ParsedLine is a ParsedCommand.
    LineCommand
)

// ParsedLine represents a parsed source line.
type ParsedLine struct {
    Row         int
    Src         []rune
    Kind        LineKind
    Label       ParsedLabel
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
    Value    string
    IsString bool
}

// Parser parses the source, and generates a sequence of ParsedInstruction's.
type Parser struct {
    lex  Tokenizer
    arch *Arch
}

func (self *Parser) err(pos int, msg string) *SyntaxError {
    return &SyntaxError {
        Pos    : pos,
        Row    : self.lex.Row,
        Src    : self.lex.Src,
        Reason : msg,
    }
}

func (self *Parser) next(p *int) rune {
    for {
        if *p >= len(self.lex.Src) {
            return 0
        } else if cc := self.lex.Src[*p]; !unicode.IsSpace(cc) {
            return cc
        } else {
            *p++
        }
    }
}

func (self *Parser) cmds() *ParsedLine {
    cmd := ""
    pos := self.lex.Pos
    buf := []ParsedCommandArg(nil)

    /* find the end of command */
    for p := pos; pos < len(self.lex.Src); pos++ {
        if unicode.IsSpace(self.lex.Src[pos]) {
            cmd = string(self.lex.Src[p:pos])
            break
        }
    }

    /* parse the arguments */
    loop: for {
        switch self.next(&pos) {
            case 0   : break loop
            case '#' : break loop
            case '"' : pos = self.strings(&buf, pos)
            default  : pos = self.expressions(&buf, pos)
        }
    }

    /* construct the line */
    return &ParsedLine {
        Row     : self.lex.Row,
        Src     : self.lex.Src,
        Kind    : LineCommand,
        Command : ParsedCommand {
            Cmd  : cmd,
            Args : buf,
        },
    }
}

func (self *Parser) init(arch *Arch) *Parser {
    self.arch = arch
    return self
}

func (self *Parser) feed(line string) *ParsedLine {
    self.lex.Row++
    self.lex.init(line)

    /* save the lexer state */
    pos := self.lex.Pos
    row := self.lex.Row

    /* parse the first token */
    tk := self.lex.Next()
    tt := tk.Ty

    /* it is a directive if it starts with a dot */
    if tk.Ty == TokenPunc && tk.Punc() == PuncDot {
        return self.cmds()
    }

    /* otherwise it could be labels or instructions */
    if tt == TokenName {
        if tkx := self.lex.Next(); tkx.Ty == TokenPunc && tkx.Punc() == PuncColon {
            tkx = self.lex.Next()
            ttx := tkx.Ty

            /* the line must end here */
            if ttx != TokenEnd {
                panic(self.err(tkx.Pos, "garbage after label definition"))
            }

            /* construct the label */
            return &ParsedLine {
                Row   : self.lex.Row,
                Src   : self.lex.Src,
                Kind  : LineLabel,
                Label : ParsedLabel {
                    Kind: Declaration,
                    Name: tk.Str,
                },
            }
        }
    }

    /* not a label, it must be an instruction */
    self.lex.Pos = pos
    self.lex.Row = row

    /* set the line kind and mnemonic */
    ret := &ParsedLine {
        Row  : self.lex.Row,
        Src  : self.lex.Src,
        Kind : LineInstr,
    }

    /* invoke the architecture specific parser */
    self.arch.impl.Parse(&self.lex, &ret.Instruction)
    return ret
}

func (self *Parser) delim(p int) int {
    if cc := self.next(&p); cc == 0 {
        return p
    } else if cc == ',' {
        return p + 1
    } else {
        panic(self.err(p, "',' expected"))
    }
}

func (self *Parser) strings(argv *[]ParsedCommandArg, p int) int {
    var i int
    var e error
    var v string

    /* find the end of string */
    for i = p + 1; i < len(self.lex.Src) && self.lex.Src[i] != '"'; i++ {
        if self.lex.Src[i] == '\\' {
            i++
        }
    }

    /* check for EOF */
    if i == len(self.lex.Src) {
        panic(self.err(i, "unexpected EOF when scanning strings"))
    }

    /* unquote the string */
    if v, e = strconv.Unquote(string(self.lex.Src[p:i + 1])); e != nil {
        panic(self.err(p, "invalid string: " + e.Error()))
    }

    /* add the argument to buffer */
    *argv = append(*argv, ParsedCommandArg { Value: v, IsString: true })
    return self.delim(i + 1)
}

func (self *Parser) directives(line string) {
    self.lex.Row++
    self.lex.init(line)

    /* parse the first token */
    tk := self.lex.Next()
    tt := tk.Ty

    /* check for EOF */
    if tt == TokenEnd {
        return
    }

    /* must be a directive */
    if tt != TokenPunc || tk.Punc() != PuncHash {
        panic(self.err(tk.Pos, "'#' expected"))
    }

    /* parse the line number */
    tk = self.lex.Next()
    tt = tk.Ty

    /* must be a line number, if it is, set the row number, and ignore the rest of the line */
    if tt != TokenInt {
        panic(self.err(tk.Pos, "line number expected"))
    } else {
        self.lex.Row = int(tk.Uint) - 1
    }
}

func (self *Parser) expressions(argv *[]ParsedCommandArg, p int) int {
    var i int
    var n int
    var s int

    /* scan until the first standalone ',' or EOF */
    loop: for i = p; i < len(self.lex.Src); i++ {
        switch self.lex.Src[i] {
            case ','           : if s == 0 { if n == 0 { break loop } }
            case ']', '}', '>' : if s == 0 { if n == 0 { break loop } else { n-- } }
            case '[', '{', '<' : if s == 0 { n++ }
            case '\\'          : if s != 0 { i++ }
            case '\''          : if s != 2 { s ^= 1 }
            case '"'           : if s != 1 { s ^= 2 }
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
    *argv = append(*argv, ParsedCommandArg { Value: string(self.lex.Src[p:i]) })
    return self.delim(i)
}

// Feed feeds the parser with one more line, and the parser
// parses it into a ParsedLine.
//
// NOTE: Feed does not handle empty lines or multiple lines,
//       it panics when this happens. Use Parse to parse multiple
//       lines of assembly source.
//
func (self *Parser) Feed(src string) (ret *ParsedLine, err error) {
    var ok bool
    var ss string
    var vv interface{}

    /* check if it is initialized */
    if self.arch == nil {
        panic("iasm: parser was not initialized properly.")
    }

    /* check for multiple lines */
    if strings.ContainsRune(src, '\n') {
        return nil, errors.New("passing multiple lines to Feed()")
    }

    /* check for blank lines */
    if ss = strings.TrimSpace(src); ss == "" || ss[0] == '#' || strings.HasPrefix(ss, "//") {
        return nil, errors.New("blank line or line with only comments or line-marks")
    }

    /* setup error handler */
    defer func() {
        if vv = recover(); vv != nil {
            if err, ok = vv.(*SyntaxError); !ok {
                panic(vv)
            }
        }
    }()

    /* call the actual parser */
    ret = self.feed(src)
    return
}

// Parse parses the entire assembly source (possibly multiple lines) into
// a sequence of *ParsedLine.
func (self *Parser) Parse(src string) (ret []*ParsedLine, err error) {
    var ok bool
    var ss string
    var vv interface{}

    /* check if it is initialized */
    if self.arch == nil {
        panic("iasm: parser was not initialized properly.")
    }

    /* setup error handler */
    defer func() {
        if vv = recover(); vv != nil {
            if err, ok = vv.(*SyntaxError); !ok {
                panic(vv)
            }
        }
    }()

    /* feed every line */
    for _, line := range strings.Split(src, "\n") {
        if ss = strings.TrimSpace(line); ss == "" || strings.HasPrefix(ss, "//") {
            self.lex.Row++
        } else if ss[0] == '#' {
            self.directives(line)
        } else {
            ret = append(ret, self.feed(line))
        }
    }

    /* everything is OK */
    err = nil
    return
}

// Directive handles the directive.
func (self *Parser) Directive(line string) (err error) {
    var ok bool
    var ss string
    var vv interface{}

    /* check if it is initialized */
    if self.arch == nil {
        panic("iasm: parser was not initialized properly.")
    }

    /* check for directives */
    if ss = strings.TrimSpace(line); ss == "" || ss[0] != '#' {
        return errors.New("not a directive")
    }

    /* setup error handler */
    defer func() {
        if vv = recover(); vv != nil {
            if err, ok = vv.(*SyntaxError); !ok {
                panic(vv)
            }
        }
    }()

    /* call the directive parser */
    self.directives(line)
    return
}
