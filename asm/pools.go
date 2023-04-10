package asm

import (
    `sync`
)

var (
    labelPool         sync.Pool
    programPool       sync.Pool
    memoryOperandPool sync.Pool
)

func freeLabel(p *Label) {
    *p = Label{}
    labelPool.Put(p)
}

// CreateLabel creates a new Label, it may allocate a new one or grab one from a pool.
func CreateLabel(name string) *Label {
    var p *Label
    var v interface{}

    /* attempt to grab from the pool */
    if v = labelPool.Get(); v != nil {
        p = v.(*Label)
    } else {
        p = new(Label)
    }

    /* initialize the label */
    p.refs = 1
    p.Name = name
    return p
}

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
    p.Arch = arch
    return p
}

func freeProgram(p *Program) {
    *p = Program{}
    programPool.Put(p)
}

func freeMemoryOperand(m *MemoryOperand) {
    *m = MemoryOperand{}
    memoryOperandPool.Put(m)
}

// CreateMemoryOperand creates a new MemoryOperand, it may allocate a new one or grab one from a pool.
func CreateMemoryOperand(ext MemoryOperandExtension) *MemoryOperand {
    var v interface{}
    var p *MemoryOperand

    /* attempt to grab from the pool */
    if v = memoryOperandPool.Get(); v != nil {
        p = v.(*MemoryOperand)
    } else {
        p = new(MemoryOperand)
    }

    /* initialize the memory operand */
    p.Ext  = ext
    p.refs = 1
    return p
}
