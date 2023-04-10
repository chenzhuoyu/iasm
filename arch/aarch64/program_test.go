package aarch64

import (
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/davecgh/go-spew/spew`
)

func TestProgram_Assemble(t *testing.T) {
    a := asm.GetArch("aarch64")
    p := Builder(a.CreateProgram())
    p.ADD(X0, X1, X2)
    spew.Dump(p.AssembleAndFree(0))
}
