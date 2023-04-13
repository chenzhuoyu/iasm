package aarch64

import (
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
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

        /* literal value */
        if tt == asm.TokenPunc && tk.Punc() == asm.PuncHash {

        }
    }
}
