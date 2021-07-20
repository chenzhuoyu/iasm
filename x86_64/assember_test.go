package x86_64

import (
    `testing`

    `github.com/davecgh/go-spew/spew`
)

func TestAssembler_Tokenizer(t *testing.T) {
    k := new(_Tokenizer)
    k.src = []rune(`movqorwhat $123, 123(%rax,%rbx), %rcx`)
    for {
        tk := k.next(true)
        if tk.tag == _T_end {
            break
        }
        spew.Dump(tk)
    }
}
