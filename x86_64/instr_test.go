package x86_64

import (
    `testing`

    `github.com/davecgh/go-spew/spew`
)

func TestEncodeInstr(t *testing.T) {
    a := CreateArch()
    p := a.CreateProgram()
    v := p.VZEROUPPER().Encode()
    spew.Dump(v)
}
