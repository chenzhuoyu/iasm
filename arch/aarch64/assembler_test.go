package aarch64

import (
    `fmt`
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/davecgh/go-spew/spew`
    `github.com/stretchr/testify/require`
)

func TestAssembler_Assemble(t *testing.T) {
    p := asm.GetArch("aarch64").CreateAssembler()
    e := p.Assemble(`
.org 0x04000000
.entry start

start:
    mov     x0, #1
    ldr     x1, =msg
    mov     x2, #14
    mov     w8, #64
    svc     #0
    mov     x0, #0
    mov     w8, #93
    svc     #0

msg:
    .ascii "Hello, ARM64!\n"
`)
    require.NoError(t, e)
    code := p.Code()
    base := p.Base()
    entry := p.Entry()
    spew.Dump(code)
    fmt.Printf("Image Base  : %#x\n", base)
    fmt.Printf("Entry Point : %#x\n", entry)
}
