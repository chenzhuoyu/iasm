package aarch64

import (
    `encoding/hex`
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/expr`
    `github.com/davecgh/go-spew/spew`
)

func TestProgram_Assemble(t *testing.T) {
    a := asm.GetArch("aarch64")
    b := asm.CreateLabel("bak")
    s := asm.CreateLabel("tab")
    j := asm.CreateLabel("jmp")
    p := Builder(a.CreateProgram())
    p.B     (j)
    p.B     (j)
    p.Link  (b)
    p.B     (b)
    p.Link  (j)
    p.ADRP  (X0, s)
    p.ADD   (X0, X0, s)
    p.LDR   (X1, Mem(X0, 5, PreIndex))
    p.ADD   (X1, X1, X0)
    p.BR    (X1)
    p.LDP   (X2, X3, Mem(X4))
    p.Align (32, expr.Int(0))
    p.Link  (s)
    p.Long  (expr.Ref(s.Retain()).Sub(expr.Ref(j.Retain())))
    p.Long  (expr.Ref(s.Retain()).Sub(expr.Ref(b.Retain())))
    ret := p.AssembleAndFree(0)
    println(hex.EncodeToString(ret))
    spew.Dump(ret)
}
