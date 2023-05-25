package obj

import (
    `debug/macho`
    `sync`
    `syscall`
)

// Arch represents the CPU architecture for object files.
type Arch struct {
    Name        string
    Align       uintptr
    MachOCpu    macho.Cpu
    MachOSubCpu uint32
}

var (
    archmut = sync.RWMutex{}
    archtab = make(map[string]*Arch)
)

// GetArch looks up the architecture by name.
func GetArch(name string) (p *Arch) {
    archmut.RLock()
    p = archtab[name]
    archmut.RUnlock()
    return
}

// RegisterArch adds a new architecture identified by its name.
func RegisterArch(arch *Arch) {
    archmut.Lock()
    archtab[arch.Name] = arch
    archmut.Unlock()
}

const (
    _CPU_SUBTYPE_LIB64   = 0x80000000
    _CPU_SUBTYPE_ARM_ALL = 0
    _CPU_SUBTYPE_X86_ALL = 3
)

var x86_64 = &Arch {
    Name        : "x86_64",
    Align       : 1,
    MachOCpu    : macho.CpuAmd64,
    MachOSubCpu : _CPU_SUBTYPE_LIB64 | _CPU_SUBTYPE_X86_ALL,
}

var aarch64 = &Arch {
    Name        : "aarch64",
    Align       : uintptr(syscall.Getpagesize()),
    MachOCpu    : macho.CpuArm64,
    MachOSubCpu : _CPU_SUBTYPE_LIB64 | _CPU_SUBTYPE_ARM_ALL,
}

func init() {
    RegisterArch(x86_64)
    RegisterArch(aarch64)
}
