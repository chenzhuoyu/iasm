package repl

import (
    `go/scanner`
    `go/token`
    `strconv`
)

type _SyntaxError string

func (self _SyntaxError) Error() string {
    return string(self)
}

type _Scanner struct {
    val string
    tok token.Token
    lex scanner.Scanner
}

const (
    dontInsertSemis scanner.Mode = 2
)

func scan(cmd string) (p *_Scanner) {
    p = new(_Scanner)
    p.lex.Init(token.NewFileSet().AddFile("(REPL)", 1, len(cmd)), []byte(cmd), nil, dontInsertSemis)
    p.next()
    return
}

func (self *_Scanner) next() {
    _, self.tok, self.val = self.lex.Scan()
}

func (self *_Scanner) must(tok token.Token) {
    if self.tok == token.EOF {
        panic(_SyntaxError("unexpected EOF"))
    } else if self.tok != tok {
        panic(_SyntaxError(tok.String() + " required"))
    }
}

func (self *_Scanner) close() {
    if self.tok != token.EOF {
        panic(_SyntaxError("excess parameters"))
    }
}

func (self *_Scanner) str(v *string) *_Scanner {
    self.must(token.STRING)
    *v, _ = strconv.Unquote(self.val)
    self.next()
    return self
}

func (self *_Scanner) uint(v *uint64) *_Scanner {
    self.must(token.INT)
    *v, _ = strconv.ParseUint(self.val, 0, 64)
    self.next()
    return self
}

func (self *_Scanner) idoff(id *uint64, offs *uint64) *_Scanner {
    self.must(token.INT)
    *id, _ = strconv.ParseUint(self.val, 0, 64)
    self.next()

    /* check for the optional offset */
    if self.tok != token.ADD {
        return self
    }

    /* parse the offset */
    self.next()
    self.must(token.INT)
    *offs, _ = strconv.ParseUint(self.val, 0, 64)
    self.next()
    return self
}

func (self *_Scanner) uintopt(v *uint64) *_Scanner {
    if self.tok != token.INT {
        return self
    } else {
        *v, _ = strconv.ParseUint(self.val, 0, 64)
        self.next()
        return self
    }
}
