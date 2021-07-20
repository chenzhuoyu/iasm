package x86_64

import (
    `fmt`
    `strconv`
    `unicode`
)

type (
	_TokenKind   int
    _Punctuation int
)

const (
    _T_end _TokenKind = iota
    _T_int
    _T_name
    _T_punc
)

const (
    _P_plus _Punctuation = iota
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

func (self _Token) String() string {
    switch self.tag {
        case _T_end  : return "<END>"
        case _T_int  : return fmt.Sprintf("<INT %d>", self.u64)
        case _T_punc : return fmt.Sprintf("<PUNC %s>", _Punctuation(self.u64))
        case _T_name : return fmt.Sprintf("<NAME %s>", strconv.QuoteToASCII(self.str))
        default      : return fmt.Sprintf("<UNK:%d %d %s>", self.tag, self.u64, strconv.QuoteToASCII(self.str))
    }
}

func tokenEnd(p int) _Token {
    return _Token {
        pos: p,
        end: p,
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

// SyntaxError represents an error in the assembly syntax.
type SyntaxError struct {
    Pos    int
    Src    string
    Reason string
}

// Error implements the error interface.
func (self *SyntaxError) Error() string {
    return self.Reason
}

type _Tokenizer struct {
    pos int
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

func (self *_Tokenizer) numx(p int, cc rune) _Token {
    if isdigit(cc) {
        return self.numv(p)
    } else if isident0(cc) {
        return tokenName(p, self.find(p, isident))
    } else {
        panic(self.err(p, "invalid char: " + strconv.QuoteRune(cc)))
    }
}

func (self *_Tokenizer) numv(p int) _Token {
    if val, err := strconv.ParseUint(self.find(p, isnumber), 0, 64); err != nil {
        panic(self.err(p, "invalid immediate value: " + err.Error()))
    } else {
        return tokenInt(p, val)
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

func (self *_Tokenizer) next(noSpace bool) _Token {
    var p int
    var c rune
    var t _Token

    /* skip spaces as needed, check for line comments */
    if noSpace {
        if self.skip(unicode.IsSpace); !self.eof() && self.ch() == '/' {
            if self.pos < len(self.src) - 1 && self.src[self.pos + 1] == '/' {
                return tokenEnd(self.pos)
            }
        }
    }

    /* check for EOF */
    if self.eof() {
        return tokenEnd(self.pos)
    }

    /* read the next character */
    p = self.pos
    c = self.rch()

    /* parse the next character */
    switch c {
        case '+' : t = tokenPunc(p, _P_plus)
        case '-' : t = tokenPunc(p, _P_minus)
        case '*' : t = tokenPunc(p, _P_star)
        case '/' : t = tokenPunc(p, _P_slash)
        case '%' : t = tokenPunc(p, _P_percent)
        case '&' : t = tokenPunc(p, _P_amp)
        case '|' : t = tokenPunc(p, _P_bar)
        case '^' : t = tokenPunc(p, _P_caret)
        case '<' : t = self.rep2(p, _P_shl, '<')
        case '>' : t = self.rep2(p, _P_shr, '>')
        case '~' : t = tokenPunc(p, _P_tilde)
        case '(' : t = tokenPunc(p, _P_lbrk)
        case ')' : t = tokenPunc(p, _P_rbrk)
        case ',' : t = tokenPunc(p, _P_comma)
        case '$' : t = tokenPunc(p, _P_dollar)
        default  : t = self.numx(p, c)
    }

    /* mark the end of token */
    t.end = self.pos
    return t
}
