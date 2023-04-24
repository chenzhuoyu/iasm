package x86_64

import (
    `fmt`
    `strings`
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
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
    require.Equal(t, asm.LineInstruction, v.Kind)
    println("Mnemonic    :", v.Instruction.Mnemonic)
    for i, x := range v.Instruction.Operands {
        println(fmt.Sprintf("Operand %3d : %s", i, x.String()))
    }
}
