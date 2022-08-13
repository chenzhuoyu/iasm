package x86_64

import (
    `fmt`
    `os`
    `os/exec`
    `strings`
    `testing`

    `github.com/chenzhuoyu/iasm/obj`
    `github.com/davecgh/go-spew/spew`
    `github.com/stretchr/testify/require`
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

func TestAssembler_Assemble(t *testing.T) {
    p := new(Assembler)
    e := p.Assemble(`
.org 0x08000000
.entry start

msg:
    .ascii "hello, world\n"

start:
    movl    $1, %edi
    leaq    msg(%rip), %rsi
    movl    $13, %edx
    movl    $0x02000004, %eax
    syscall
    xorl    %edi, %edi
    movl    $0x02000001, %eax
    syscall
`)
    require.NoError(t, e)
    code := p.Code()
    base := p.Base()
    entry := p.Entry()
    spew.Dump(code)
    fmt.Printf("Image Base  : %#x\n", base)
    fmt.Printf("Entry Point : %#x\n", entry)
    fp, err := os.CreateTemp("", "macho_out-")
    require.NoError(t, err)
    err = obj.MachO.Write(fp, code, uint64(base), uint64(entry))
    require.NoError(t, err)
    err = fp.Close()
    require.NoError(t, err)
    err = os.Chmod(fp.Name(), 0755)
    require.NoError(t, err)
    println("Saved to", fp.Name())
    out, err := exec.Command(fp.Name()).Output()
    require.NoError(t, err)
    spew.Dump(out)
    require.Equal(t, []byte("hello, world\n"), out)
    err = os.Remove(fp.Name())
    require.NoError(t, err)
}

func TestAssembler_RIPRelative(t *testing.T) {
    p := new(Assembler)
    e := p.Assemble(`leaq 0x1b(%rip), %rsi`)
    require.NoError(t, e)
    spew.Dump(p.Code())
    require.Equal(t, []byte { 0x48, 0x8d, 0x35, 0x1b, 0x00, 0x00, 0x00 }, p.Code())
}
