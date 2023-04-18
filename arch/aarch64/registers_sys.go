package aarch64

import (
    `fmt`

    `github.com/chenzhuoyu/iasm/asm`
)

// SystemRegister represents AArch64 system registers.
type SystemRegister uint8

var _SysRegisterNames = [256]string {
    // TODO: this
}

func (self SystemRegister) ID()     uint8  { return uint8(self) }
func (self SystemRegister) String() string { return _SysRegisterNames[self] }

var (
	_SysRegisters = make(map[string]asm.Register)
)

func init() {
    for i, name := range _SysRegisterNames {
        if name != "" {
            _SysRegisters[name] = SystemRegister(i)
        } else {
            _SysRegisterNames[i] = fmt.Sprintf("{SYSREG %#x}", i)
        }
    }
}
