package x86_64

import (
    `sync`
)

var (
    programPool       sync.Pool
    instructionPool   sync.Pool
    memoryOperandPool sync.Pool
)

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
        arch   : arch,
        instrs : make([]*Instruction, 0, 16),
    }
}

func resetProgram(arch *Arch, prog *Program) *Program {
    prog.arch = arch
    prog.instrs = prog.instrs[:0]
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
        len  : 0,
        argc : argc,
        argv : argv,
    }
}

func resetInstruction(argc int, argv Operands, instr *Instruction) *Instruction {
    instr.len = 0
    instr.argc = argc
    instr.argv = argv
    return instr
}

func newMemoryOperand() *MemoryOperand {
    if v := memoryOperandPool.Get(); v == nil {
        return new(MemoryOperand)
    } else {
        return resetMemoryOperand(v.(*MemoryOperand))
    }
}

func freeMemoryOperand(m interface{}) {
    if _, ok := m.(*MemoryOperand); ok {
        memoryOperandPool.Put(m)
    }
}

func resetMemoryOperand(m *MemoryOperand) *MemoryOperand {
    *m = MemoryOperand{}
    return m
}
