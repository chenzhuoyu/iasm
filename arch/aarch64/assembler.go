package aarch64

import (
    `unsafe`

    `github.com/chenzhuoyu/iasm/asm`
)

type _AssemblerImpl struct {
    asm.Assembler
}

func asmthis(p *asm.Assembler) *_AssemblerImpl {
    return (*_AssemblerImpl)(unsafe.Pointer(p))
}

func (self *_AssemblerImpl) build(p *Program, line *asm.ParsedInstruction) (err error) {
    // TODO: this
    // panic("aarch64: _AssemblerImpl.build(): not implemented")
    println(line.String())
    return nil
}
