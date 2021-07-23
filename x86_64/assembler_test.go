package x86_64

import (
    `strings`
    `testing`

    `github.com/davecgh/go-spew/spew`
)

func TestAssembler_Tokenizer(t *testing.T) {
    k := new(_Tokenizer)
    k.src = []rune(`movqorwhat $123, 123(%rax,%rbx), %rcx`)
    for {
        tk := k.next()
        if tk.tag == _T_end {
            break
        }
        spew.Dump(tk)
    }
}

func TestAssembler_Parser(t *testing.T) {
    p := new(Parser)
    v := p.Feed("movq " + strings.Join([]string {
        `$123`,
        `$-123`,
        `%rcx`,
        `(%rax)`,
        `123(%rax)`,
        `-123(%rax)`,
        `(%rax,%rbx,4)`,
        `1234(%rax,%rbx,4)`,
        `(,%rax,8)`,
        `1234(,%rax,8)`,
        `$(123 + 456)`,
        `(123 + 456)(%rax)`,
        `$'asd'`,
        `$'asdf'`,
        `$'asdfghj'`,
        `$'asdfghjk'`,
    }, ", "))
    spew.Dump(v)
}
