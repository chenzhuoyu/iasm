package x86_64

import (
    `bytes`
    `testing`

    `github.com/chenzhuoyu/iasm/expr`
    `github.com/davecgh/go-spew/spew`
)

func TestProgram_Assemble(t *testing.T) {
    a := CreateArch()
    b := CreateLabel("bak")
    s := CreateLabel("tab")
    j := CreateLabel("jmp")
    p := a.CreateProgram()
    p.JMP    (j)
    p.JMP    (j)
    p.Link   (b)
    p.Data   (bytes.Repeat([]byte { 0x0f, 0x1f, 0x84, 0x00, 0x00, 0x00, 0x00, 0x00 }, 15))
    p.Data   ([]byte { 0x0f, 0x1f, 0x00 })
    p.JMP    (b)
    p.Link   (j)
    p.LEAQ   (Ref(s), RDI)
    p.MOVSLQ (Sib(RDI, RAX, 4, -4), RAX)
    p.ADDQ   (RDI, RAX)
    p.JMPQ   (RAX)
    p.Align  (32, expr.Int(0xcc))
    p.Link   (s)
    p.Long   (expr.Ref(s.Retain()).Sub(expr.Ref(j.Retain())))
    p.Long   (expr.Ref(s.Retain()).Sub(expr.Ref(b.Retain())))
    spew.Dump(p.AssembleAndFree(0))
}
