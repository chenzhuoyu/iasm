package x86_64

import (
    `strings`
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/davecgh/go-spew/spew`
    `github.com/stretchr/testify/require`
)

func TestParser_Parse(t *testing.T) {
    p := asm.GetArch("x86_64").CreateParser()
    v, e := p.Feed("movq " + strings.Join([]string {
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
    require.NoError(t, e)
    spew.Dump(v)
}
