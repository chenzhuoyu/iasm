package aarch64

import (
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/davecgh/go-spew/spew`
)

func TestProgram_Assemble(t *testing.T) {
    a := asm.GetArch("aarch64")
    p := a.CreateProgram().(*Program)
    p.ADD(X0, SP, X2)
    spew.Dump(asm.AssembleAndFree(p, 0))
}
