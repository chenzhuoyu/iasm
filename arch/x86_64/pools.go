package x86_64

import (
    `sync`
)

var (
    programPool                sync.Pool
    instructionPool            sync.Pool
    memoryOperandExtensionPool sync.Pool
)

func newProgram(arch *Arch) *Program {
    var p *Program
    var v interface{}

    /* attempt to grab from the pool */
    if v = programPool.Get(); v != nil {
        p = v.(*Program)
    } else {
        p = new(Program)
    }

    /* initialize the program */
    p.arch = arch
    return p
}

func freeProgram(p *Program) {
    *p = Program{}
    programPool.Put(p)
}

func newInstruction(name string, argc int, argv Operands) *Instruction {
    var v interface{}
    var p *Instruction

    /* attempt to grab from the pool */
    if v = instructionPool.Get(); v != nil {
        p = v.(*Instruction)
    } else {
        p = new(Instruction)
    }

    /* initialize the instruction */
    p.refs = 1
    p.name = name
    p.argc = argc
    p.argv = argv
    return p
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
