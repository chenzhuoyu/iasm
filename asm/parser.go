package asm

import (
    `errors`
    `fmt`
    `reflect`
    `strconv`
    `strings`
    `unicode`

    `github.com/chenzhuoyu/iasm/expr`
)

// TokenKind represents the type of a token.
type TokenKind int

const (
    TokenEnd TokenKind = iota + 1
    TokenInt
    TokenFp64
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
    PuncLBrace
    PuncRBrace
    PuncLIndex
    PuncRIndex
    PuncLCurly
    PuncRCurly
    PuncDot
    PuncComma
    PuncColon
    PuncDollar
    PuncHash
    PuncBang
    PuncEqual
)

var _PuncNames = map[Punctuation]string {
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
    PuncLBrace  : "(",
    PuncRBrace  : ")",
    PuncLIndex  : "[",
    PuncRIndex  : "]",
    PuncLCurly  : "{",
    PuncRCurly  : "}",
    PuncDot     : ".",
    PuncComma   : ",",
    PuncColon   : ":",
    PuncDollar  : "$",
    PuncHash    : "#",
    PuncBang    : "!",
    PuncEqual   : "=",
}

func (self Punctuation) String() string {
    if v, ok := _PuncNames[self]; ok {
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
    Fp64 float64
}

// Punc converts the token into a Punctuation.
func (self Token) Punc() Punctuation {
    if self.Ty != TokenPunc {
        panic("asm: token is not a punctuation")
    } else {
        return Punctuation(self.Uint)
    }
}

func (self Token) String() string {
    switch self.Ty {
        case TokenEnd   : return "<END>"
        case TokenInt   : return fmt.Sprintf("<INT %d>", self.Uint)
        case TokenFp64  : return fmt.Sprintf("<FP64 %f>", self.Fp64)
        case TokenPunc  : return fmt.Sprintf("<PUNC %s>", Punctuation(self.Uint))
        case TokenName  : return fmt.Sprintf("<NAME %s>", strconv.QuoteToASCII(self.Str))
        case TokenSpace : return "<SPACE>"
        default         : return fmt.Sprintf("<UNK:%d %d %s>", self.Ty, self.Uint, strconv.QuoteToASCII(self.Str))
    }
}

// IsPunc tests if this token matches the specified Punctuation.
func (self Token) IsPunc(punc Punctuation) bool {
    return self.Ty == TokenPunc && Punctuation(self.Uint) == punc
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

func tokenFp64(p int, val float64) Token {
    return Token {
        Ty   : TokenFp64,
        Pos  : p,
        Fp64 : val,
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
        return fmt.Sprintf("Syntax error at line %d: %s", self.Row, self.Reason)
    } else {
        return fmt.Sprintf("Syntax error at [%d:%d]: %s", self.Row, self.Pos + 1, self.Reason)
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
    ret, self.Pos = self.Src[self.Pos], self.Pos + 1
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
        if self.Pos++; self.Src[self.Pos - 1] == '\n' {
            self.Row++
        }
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
    if num := self.find(p, isnumber); num == "" {
        panic(self.err(p, "missing immediate value"))
    } else if val, err := strconv.ParseUint(num, 0, 64); err == nil {
        return tokenInt(p, val)
    } else if f64, err := strconv.ParseFloat(num, 64); err == nil {
        return tokenFp64(p, f64)
    } else {
        panic(self.err(p, "invalid immediate value: " + err.Error()))
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

func (self *Tokenizer) evalfn(p expr.Repository, fn func(*expr.Parser, expr.Repository) (*expr.Expr, error)) int64 {
    var err error
    var ret int64
    var ast *expr.Expr
    var exp expr.Parser

    /* skip all the spaces */
    if unicode.IsSpace(self.Src[self.Pos]) {
        self.skip(unicode.IsSpace)
    }

    /* save the current location */
    pos := self.Pos
    row := self.Row

    /* special case: character literal */
    if !self.eof() && self.ch() == '\'' {
        return int64(self.chrv(pos).Uint)
    }

    /* parse the expression */
    if ast, err = fn(exp.SetSource(self.Src[pos:]), p); err != nil {
        if s, ok := err.(*expr.SyntaxError); ok {
            panic(self.err(pos + s.Pos, s.Reason))
        } else {
            panic(self.err(pos + exp.Pos(), "cannot parse expression: " + err.Error()))
        }
    }

    /* evaluate the expression */
    ret, err = ast.Evaluate()
    ast.Free()

    /* check for errors */
    if err != nil {
        if s, ok := err.(*expr.SyntaxError); ok {
            panic(self.err(pos + s.Pos, s.Reason))
        } else {
            panic(self.err(pos + exp.Pos(), "cannot evaluate expression: " + err.Error()))
        }
    }

    /* update tokenizer location */
    self.Row = row + exp.Row()
    self.Pos = pos + exp.Pos()
    return ret
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
        case '('  : t = tokenPunc(p, PuncLBrace)
        case ')'  : t = tokenPunc(p, PuncRBrace)
        case '['  : t = tokenPunc(p, PuncLIndex)
        case ']'  : t = tokenPunc(p, PuncRIndex)
        case '{'  : t = tokenPunc(p, PuncLCurly)
        case '}'  : t = tokenPunc(p, PuncRCurly)
        case '.'  : t = tokenPunc(p, PuncDot)
        case ','  : t = tokenPunc(p, PuncComma)
        case ':'  : t = tokenPunc(p, PuncColon)
        case '$'  : t = tokenPunc(p, PuncDollar)
        case '#'  : t = tokenPunc(p, PuncHash)
        case '!'  : t = tokenPunc(p, PuncBang)
        case '='  : t = tokenPunc(p, PuncEqual)
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

// Eval evaluates the expression at current location, advancing the tokenizer.
func (self *Tokenizer) Eval(p expr.Repository) int64 {
    return self.evalfn(p, (*expr.Parser).Parse)
}

// Value is like Eval, but it only evaluates a single term rather than the whole expression.
func (self *Tokenizer) Value(p expr.Repository) int64 {
    return self.evalfn(p, (*expr.Parser).Value)
}

// MatchEnd attempts to matche the TokenEnd.
func (self *Tokenizer) MatchEnd() bool {
    pos := self.Pos
    row := self.Row
    tok := self.Next()

    /* check for the punctuation */
    if tok.Ty == TokenEnd {
        return true
    }

    /* otherwise rewind the tokenizer */
    self.Pos = pos
    self.Row = row
    return false
}

// MatchPunc attempts to matche the next token for a specified Punctuation.
// Consumes it if matches.
func (self *Tokenizer) MatchPunc(punc Punctuation) bool {
    pos := self.Pos
    row := self.Row
    tok := self.Next()

    /* check for the punctuation */
    if tok.IsPunc(punc) {
        return true
    }

    /* otherwise rewind the tokenizer */
    self.Pos = pos
    self.Row = row
    return false
}

// MatchPuncNow attempts to matche the next adjacent token (including whitespaces) for a specified Punctuation.
// Consumes it if matches.
func (self *Tokenizer) MatchPuncNow(punc Punctuation) bool {
    pos := self.Pos
    row := self.Row
    tok := self.Read()

    /* check for the punctuation */
    if tok.IsPunc(punc) {
        return true
    }

    /* otherwise rewind the tokenizer */
    self.Pos = pos
    self.Row = row
    return false
}

// ExpectPunc matches the next token for a specified Punctuation.
// Throws an error if not match.
func (self *Tokenizer) ExpectPunc(punc Punctuation) {
    if tk := self.Next(); !tk.IsPunc(punc) {
        panic(self.err(tk.Pos, fmt.Sprintf("%q expected", punc.String())))
    }
}

// ExpectPuncNow matches the next adjacent token (including whitespaces) for a specified Punctuation.
// Throws an error if not match.
func (self *Tokenizer) ExpectPuncNow(punc Punctuation) {
    if tk := self.Read(); !tk.IsPunc(punc) {
        panic(self.err(tk.Pos, fmt.Sprintf("%q expected", punc.String())))
    }
}

// OperandKind indicates the type of the operand.
type OperandKind int

const (
    // OpImm means the operand is an immediate value.
    OpImm OperandKind = iota

    // OpFpImm means the operand is a floating-point immediate value.
    OpFpImm

    // OpReg means the operand is a register.
    OpReg

    // OpMem means the operand is a memory address.
    OpMem

    // OpMod means the operand is an address modifier.
    OpMod

    // OpSym means the operand is a literal symbol.
    OpSym

    // OpLabel means the operand is a label, specifically for
    // branch instructions.
    OpLabel
)

func (self OperandKind) String() string {
    switch self {
        case OpImm   : return "Imm"
        case OpFpImm : return "FpImm"
        case OpReg   : return "Reg"
        case OpMem   : return "Mem"
        case OpMod   : return "Mod"
        case OpSym   : return "Sym"
        case OpLabel : return "Label"
        default      : return "???"
    }
}

// LabelKind indicates the type of label reference.
type LabelKind int

const (
    // Declaration means the label is a declaration.
    Declaration LabelKind = iota + 1

    // BranchTarget means the label should be treated as a branch target.
    BranchTarget

    // RelativeAddress means the label should be treated as a reference to
    // the code section (e.g. RIP-relative addressing).
    RelativeAddress
)

func (self LabelKind) String() string {
    switch self {
        case Declaration     : return "Declaration"
        case BranchTarget    : return "BranchTarget"
        case RelativeAddress : return "RelativeAddress"
        default              : return "???"
    }
}

// InstructionPrefix indicates the prefix bytes prepended to the instruction.
type InstructionPrefix interface {
    fmt.Stringer
    EncodePrefix(p *Instruction)
}

// ParsedLabel represents a label in the source, either a jump target or
// an RIP-relative addressing.
type ParsedLabel struct {
    Name string
    Kind LabelKind
}

func (self *ParsedLabel) String() string {
    return fmt.Sprintf("(%s) %s", self.Kind, self.Name)
}

// ParsedSymbol represents an operand of a literal symbol in the source.
type ParsedSymbol struct {
    Name  string
    Value interface{}
}

func (self *ParsedSymbol) String() string {
    return fmt.Sprintf("%s: %s = %v", self.Name, reflect.TypeOf(self.Value), self.Value)
}

// ParsedModifier represents an operand of an address modifier in the source.
type ParsedModifier struct {
    Value interface{}
}

func (self *ParsedModifier) String() string {
    return fmt.Sprint(self.Value)
}

// ParsedOperand represents an operand of an instruction in the source.
type ParsedOperand struct {
    Op    OperandKind
    Imm   int64
    FpImm float64
    Reg   Register
    Mem   MemoryAddress
    Mod   ParsedModifier
    Sym   ParsedSymbol
    Label ParsedLabel
}

func (self *ParsedOperand) String() string {
    switch self.Op {
        case OpImm   : return fmt.Sprintf("(%s) %#x", self.Op, self.Imm)
        case OpFpImm : return fmt.Sprintf("(%s) %f", self.Op, self.FpImm)
        case OpReg   : return fmt.Sprintf("(%s) %s", self.Op, self.Reg)
        case OpMem   : return fmt.Sprintf("(%s) %s", self.Op, self.Mem.String())
        case OpMod   : return fmt.Sprintf("(%s) %s", self.Op, self.Mod.String())
        case OpSym   : return fmt.Sprintf("(%s) %s", self.Op, self.Sym.String())
        case OpLabel : return fmt.Sprintf("(%s) %s", self.Op, self.Label.String())
        default      : return "???"
    }
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
        Op  : OpMem,
        Mem : v,
    })
}

// Mod adds an address modifier operand to this instruction.
func (self *ParsedInstruction) Mod(mod interface{}) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op  : OpMod,
        Mod : ParsedModifier { mod },
    })
}

// Sym adds a literal symbol operand to this instruction.
func (self *ParsedInstruction) Sym(name string, value interface{}) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op  : OpSym,
        Sym : ParsedSymbol {
            Name  : name,
            Value : value,
        },
    })
}

// FpImm adds a floating-point immediate operand to this instruction.
func (self *ParsedInstruction) FpImm(v float64) {
    self.Operands = append(self.Operands, ParsedOperand {
        Op    : OpFpImm,
        FpImm : v,
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

func (self *ParsedInstruction) String() string {
    pfx := make([]string, 0, len(self.Prefixes))
    ops := make([]string, 0, len(self.Operands))

    /* add prefixes */
    for _, v := range self.Prefixes {
        pfx = append(pfx, v.String())
    }

    /* add operands */
    for _, v := range self.Operands {
        ops = append(ops, v.String())
    }

    /* add mnemonic and operands */
    if pfx = append(pfx, self.Mnemonic); len(ops) == 0 {
        return strings.Join(pfx, " ")
    } else {
        return strings.Join(append(pfx, strings.Join(ops, ", ")), " ")
    }
}

// LineKind indicates the type of ParsedLine.
type LineKind int

const (
    // LineLabel means the ParsedLine is a label.
    LineLabel LineKind = iota + 1

    // LineCommand means the ParsedLine is a ParsedCommand.
    LineCommand

    // LineInstruction means the ParsedLine is an instruction.
    LineInstruction

)

func (self LineKind) String() string {
    switch self {
        case LineLabel       : return "Label"
        case LineCommand     : return "Command"
        case LineInstruction : return "Instruction"
        default              : return "???"
    }
}

// ParsedLine represents a parsed source line.
type ParsedLine struct {
    Row         int
    Src         []rune
    Kind        LineKind
    Label       ParsedLabel
    Command     ParsedCommand
    Instruction ParsedInstruction
}

func (self *ParsedLine) String() string {
    switch self.Kind {
        case LineLabel       : return fmt.Sprintf("(%s) %s", self.Kind, self.Label.String())
        case LineCommand     : return fmt.Sprintf("(%s) %s", self.Kind, self.Command.String())
        case LineInstruction : return fmt.Sprintf("(%s) %s", self.Kind, self.Instruction.String())
        default              : return "???"
    }
}

// ParsedCommand represents a parsed assembly directive command.
type ParsedCommand struct {
    Cmd  string
    Args []ParsedCommandArg
}

func (self *ParsedCommand) String() string {
    nb := len(self.Args)
    ret := make([]string, 0, nb)

    /* add arguments */
    for _, v := range self.Args {
        ret = append(ret, v.String())
    }

    /* join them together */
    if nb == 0 {
        return self.Cmd
    } else {
        return fmt.Sprintf("%s %s", self.Cmd, strings.Join(ret, ", "))
    }
}

// ParsedCommandArg represents an argument of a ParsedCommand.
type ParsedCommandArg struct {
    Value    string
    IsString bool
}

func (self *ParsedCommandArg) String() string {
    if !self.IsString {
        return self.Value
    } else {
        return strconv.Quote(self.Value)
    }
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

    /* parse the first token */
    pos := self.lex.Pos
    row := self.lex.Row
    tok := self.lex.Next()

    /* it is a directive if it starts with a dot */
    if tok.IsPunc(PuncDot) {
        return self.cmds()
    }

    /* otherwise it must be an identifier (either instruction mnemonic or label name) */
    if tok.Ty != TokenName {
        panic(self.err(tok.Pos, "'.' or identifier expected"))
    }

    /* label must be in the form of "<ident>:" */
    if self.lex.Next().IsPunc(PuncColon) {
        if self.lex.Next().Ty == TokenEnd {
            return &ParsedLine {
                Row   : self.lex.Row,
                Src   : self.lex.Src,
                Kind  : LineLabel,
                Label : ParsedLabel {
                    Kind: Declaration,
                    Name: tok.Str,
                },
            }
        }
    }

    /* not a label, it must be an instruction, revert the tokenizer */
    self.lex.Pos = pos
    self.lex.Row = row

    /* set the line kind and mnemonic */
    ret := &ParsedLine {
        Row  : self.lex.Row,
        Src  : self.lex.Src,
        Kind : LineInstruction,
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
    self.lex.init(line)
    self.lex.ExpectPunc(PuncHash)
    self.lex.Row = int(self.lex.Value(nil)) - 1
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
