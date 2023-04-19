package aarch64

import (
    `fmt`
)

const (
    _SR_read  = 0b10 << _SR_rwoffs
    _SR_write = 0b01 << _SR_rwoffs
)

// SystemRegister represents AArch64 system registers.
type SystemRegister uint32

// Sr composes a read-write system register with "11:op1:CRn:CRm:op2".
func Sr(op1, cn, cm, op2 uint32) SystemRegister {
    if op1 &^ 0x07 != 0 || cn &^ 0x0f != 0 || cm &^ 0x0f != 0 || op2 &^ 0x07 != 0 {
        panic("aarch64: invalid system register ID")
    } else {
        return SystemRegister((0b11 << _SR_rwoffs) | (cm << 11) | (cn << 7) | (1 << 6) | (op1 << 3) | op2)
    }
}

func (self SystemRegister) r() uint32 {
    if self & _SR_read == 0 {
        panic(fmt.Sprintf("aarch64: System Register %s is not readable.", self.String()))
    } else {
        return uint32(self & _SR_regmask)
    }
}

func (self SystemRegister) w() uint32 {
    if self & _SR_write == 0 {
        panic(fmt.Sprintf("aarch64: System Register %s is not writable.", self.String()))
    } else {
        return uint32(self & _SR_regmask)
    }
}

func (self SystemRegister) ID() uint8 {
    panic("aarch64: invalid use of System Registers.")
}

func (self SystemRegister) String() string {
    if ret, ok := _SysRegisterNames[self]; ok {
        return ret
    } else {
        return "???"
    }
}
