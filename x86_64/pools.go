package x86_64

import (
    `sync`
)

var (
    labelPool         sync.Pool
    programPool       sync.Pool
    instructionPool   sync.Pool
    memoryOperandPool sync.Pool
)

func freeLabel(v interface{}) {
    if _, ok := v.(*Label); ok {
        labelPool.Put(v)
    }
}

func allocLabel(name string) *Label {
    return &Label {
        Dest: nil,
        Name: name,
    }
}

func resetLabel(name string, label *Label) *Label {
    label.Dest = nil
    label.Name = name
    return label
}

// CreateLabel creates a new Label, it may allocate a new one or grab one from a pool.
func CreateLabel(name string) *Label {
    if p := labelPool.Get(); p == nil {
        return allocLabel(name)
    } else {
        return resetLabel(name, p.(*Label))
    }
}

func newProgram(arch *Arch) *Program {
    if p := programPool.Get(); p == nil {
        return allocProgram(arch)
    } else {
        return resetProgram(arch, p.(*Program))
    }
}

func freeProgram(p *Program) {
    programPool.Put(p)
}

func allocProgram(arch *Arch) *Program {
    return &Program {
        arch: arch,
        head: nil,
        tail: nil,
    }
}

func resetProgram(arch *Arch, prog *Program) *Program {
    prog.head = nil
    prog.tail = nil
    prog.arch = arch
    return prog
}

func newInstruction(argc int, argv Operands) *Instruction {
    if v := instructionPool.Get(); v == nil {
        return allocInstruction(argc, argv)
    } else {
        return resetInstruction(argc, argv, v.(*Instruction))
    }
}

func freeInstruction(v *Instruction) {
    instructionPool.Put(v)
}

func allocInstruction(argc int, argv Operands) *Instruction {
    return &Instruction {
        argc : argc,
        argv : argv,
    }
}

func resetInstruction(argc int, argv Operands, instr *Instruction) *Instruction {
    instr.pc = 0
    instr.nb = 0
    instr.len = 0
    instr.argc = argc
    instr.argv = argv
    instr.branch = false
    instr.pseudo = _Pseudo{}
    return instr
}

func freeMemoryOperand(m interface{}) {
    if _, ok := m.(*MemoryOperand); ok {
        memoryOperandPool.Put(m)
    }
}

// CreateMemoryOperand creates a new MemoryOperand, it may allocate a new one or grab one from a pool.
func CreateMemoryOperand() *MemoryOperand {
    var v interface{}
    var m *MemoryOperand

    /* attempt to grab from the pool */
    if v = memoryOperandPool.Get(); v == nil {
        return new(MemoryOperand)
    }

    /* clear and reuse the operand */
    m = v.(*MemoryOperand)
    *m = MemoryOperand{}
    return m
}
