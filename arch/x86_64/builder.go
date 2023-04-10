package x86_64

import (
    `unsafe`

    `github.com/chenzhuoyu/iasm/asm`
)

// Builder returns the x86_64 specific instruction builder.
func Builder(p *asm.Program) *Program {
    return (*Program)(unsafe.Pointer(p))
}
