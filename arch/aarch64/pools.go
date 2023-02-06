package aarch64

import (
    `sync`
)

var (
    programPool     sync.Pool
    instructionPool sync.Pool
)

func newProgram() *Program {
    if v := programPool.Get(); v != nil {
        return v.(*Program)
    } else {
        return new(Program)
    }
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
    *p = Instruction{}
    instructionPool.Put(p)
}
