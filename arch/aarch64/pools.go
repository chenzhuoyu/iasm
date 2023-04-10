package aarch64

import (
    `sync`
)

var (
    instructionPool sync.Pool
)

func newInstruction() *Instruction {
    if v := instructionPool.Get(); v != nil {
        return v.(*Instruction)
    } else {
        return new(Instruction)
    }
}

func freeInstruction(p *Instruction) {
    *p = Instruction{}
    instructionPool.Put(p)
}
