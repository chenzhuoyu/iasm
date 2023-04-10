package x86_64

import (
    `github.com/chenzhuoyu/iasm/asm`
)

type (
	_ArchImpl struct{}
)

func (_ArchImpl) New      () *asm.Instruction                 { return &newInstruction().Instruction }
func (_ArchImpl) Free     (p *asm.Instruction)                { freeInstruction(this(p)) }
func (_ArchImpl) Encode   (p *asm.Instruction, m *[]byte) int { return this(p).encode(m) }
func (_ArchImpl) Assemble (p *asm.Program, pc uintptr) []byte { return Builder(p).assemble(pc) }

func init() {
    asm.RegisterArch("x86_64", _FEAT_MAX, new(_ArchImpl))
}
