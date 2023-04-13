package asm

import (
    `testing`

    `github.com/davecgh/go-spew/spew`
)

func TestAssembler_Tokenizer(t *testing.T) {
    k := new(Tokenizer)
    k.Src = []rune(`movqorwhat $123, 123(%rax,%rbx), %rcx`)
    for {
        tk := k.Next()
        if tk.Ty == TokenEnd {
            break
        }
        spew.Dump(tk)
    }
}
