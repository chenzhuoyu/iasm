package x86_64

import (
    `sync`
)

var (
    instructionPool            sync.Pool
    memoryOperandExtensionPool sync.Pool
)

func newInstruction() *Instruction {
    if v := instructionPool.Get(); v != nil {
        return v.(*Instruction)
    } else {
        return new(Instruction)
    }
}

func freeInstruction(p *Instruction) {
    *p = Instruction { prefix: p.prefix[:0] }
    instructionPool.Put(p)
}

func newMemoryOperandExtension() *MemoryOperandExtension {
    if v := memoryOperandExtensionPool.Get(); v != nil {
        return v.(*MemoryOperandExtension)
    } else {
        return new(MemoryOperandExtension)
    }
}

func freeMemoryOperandExtension(p *MemoryOperandExtension) {
    *p = MemoryOperandExtension{}
    memoryOperandExtensionPool.Put(p)
}
