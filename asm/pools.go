package asm

import (
    `sync`
)

var (
    labelPool         sync.Pool
    memoryOperandPool sync.Pool
)

func freeLabel(v *Label) {
    labelPool.Put(v)
}

func clearLabel(p *Label) *Label {
    *p = Label{}
    return p
}

// CreateLabel creates a new Label, it may allocate a new one or grab one from a pool.
func CreateLabel(name string) *Label {
    var p *Label
    var v interface{}

    /* attempt to grab from the pool */
    if v = labelPool.Get(); v == nil {
        p = new(Label)
    } else {
        p = clearLabel(v.(*Label))
    }

    /* initialize the label */
    p.refs = 1
    p.Name = name
    return p
}

func freeMemoryOperand(m *MemoryOperand) {
    memoryOperandPool.Put(m)
}

func clearMemoryOperand(m *MemoryOperand) *MemoryOperand {
    *m = MemoryOperand{}
    return m
}

// CreateMemoryOperand creates a new MemoryOperand, it may allocate a new one or grab one from a pool.
func CreateMemoryOperand(ext MemoryOperandExtension) *MemoryOperand {
    var v interface{}
    var p *MemoryOperand

    /* attempt to grab from the pool */
    if v = memoryOperandPool.Get(); v == nil {
        p = new(MemoryOperand)
    } else {
        p = clearMemoryOperand(v.(*MemoryOperand))
    }

    /* initialize the memory operand */
    p.Ext  = ext
    p.refs = 1
    return p
}
