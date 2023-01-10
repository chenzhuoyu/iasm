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
    if v = programPool.Get(); v == nil {
        p = new(Program)
    } else {
        p = clearProgram(v.(*Program))
    }

    /* initialize the program */
    p.arch = arch
    return p
}

func freeProgram(p *Program) {
    programPool.Put(p)
}

func clearProgram(p *Program) *Program {
    *p = Program{}
    return p
}

func newInstruction(name string, argc int, argv Operands) *Instruction {
    var v interface{}
    var p *Instruction

    /* attempt to grab from the pool */
    if v = instructionPool.Get(); v == nil {
        p = new(Instruction)
    } else {
        p = clearInstruction(v.(*Instruction))
    }

    /* initialize the instruction */
    p.name = name
    p.argc = argc
    p.argv = argv
    return p
}

func freeInstruction(v *Instruction) {
    instructionPool.Put(v)
}

func clearInstruction(p *Instruction) *Instruction {
    *p = Instruction { prefix: p.prefix[:0] }
    return p
}

func newMemoryOperandExtension() *MemoryOperandExtension {
    if v := memoryOperandExtensionPool.Get(); v == nil {
        return new(MemoryOperandExtension)
    } else {
        return clearMemoryOperandExtension(v.(*MemoryOperandExtension))
    }
}

func freeMemoryOperandExtension(p *MemoryOperandExtension) {
    memoryOperandExtensionPool.Put(p)
}

func clearMemoryOperandExtension(p *MemoryOperandExtension) *MemoryOperandExtension {
    *p = MemoryOperandExtension{}
    return p
}
