package repl

import (
    `fmt`

    `github.com/klauspost/cpuid/v2`
)

type _RegFileAMD64 struct {
    rax    uint64
    rbx    uint64
    rcx    uint64
    rdx    uint64
    rdi    uint64
    rsi    uint64
    rbp    uint64
    rsp    uint64
    r8     uint64
    r9     uint64
    r10    uint64
    r11    uint64
    r12    uint64
    r13    uint64
    r14    uint64
    r15    uint64
    rip    uint64
    rflags uint64
    cs     uint16
    _      [6]byte
    fs     uint16
    _      [6]byte
    gs     uint16
    _      [30]byte
    xmm    [32][16]byte
    ymm    [32][32]byte
    zmm    [32][64]byte
    k      [8]uint64
}

var (
    hasAVX      = cpuid.CPU.Has(cpuid.AVX)
    hasAVX512F  = cpuid.CPU.Has(cpuid.AVX512F)
    hasAVX512BW = cpuid.CPU.Has(cpuid.AVX512BW)
    hasAVX512VL = cpuid.CPU.Has(cpuid.AVX512VL)
)

//go:nosplit
//go:noescape
//goland:noinspection GoUnusedParameter
func dumpregs(regs *_RegFileAMD64)

func (self *_RegFileAMD64) Dump(indent int) []_Reg {
    dumpregs(self)
    return self.format(indent)
}

func (self *_RegFileAMD64) Compare(other _RegFile, indent int) (v []_Reg) {
    var ok bool
    var rf *_RegFileAMD64

    /* must be an AMD64 register file */
    if rf, ok = other.(*_RegFileAMD64); !ok {
        panic("can only compare with AMD64 register file")
    }

    /* compare generic registers, %rip is always going
     * to change, so ignore this in the comparison result */
    if self.rax    != rf.rax    { v = append(v, _Reg{"rax"   , fmt.Sprintf("%#016x -> %#016x", self.rax   , rf.rax   ), false}) }
    if self.rbx    != rf.rbx    { v = append(v, _Reg{"rbx"   , fmt.Sprintf("%#016x -> %#016x", self.rbx   , rf.rbx   ), false}) }
    if self.rcx    != rf.rcx    { v = append(v, _Reg{"rcx"   , fmt.Sprintf("%#016x -> %#016x", self.rcx   , rf.rcx   ), false}) }
    if self.rdx    != rf.rdx    { v = append(v, _Reg{"rdx"   , fmt.Sprintf("%#016x -> %#016x", self.rdx   , rf.rdx   ), false}) }
    if self.rdi    != rf.rdi    { v = append(v, _Reg{"rdi"   , fmt.Sprintf("%#016x -> %#016x", self.rdi   , rf.rdi   ), false}) }
    if self.rsi    != rf.rsi    { v = append(v, _Reg{"rsi"   , fmt.Sprintf("%#016x -> %#016x", self.rsi   , rf.rsi   ), false}) }
    if self.rbp    != rf.rbp    { v = append(v, _Reg{"rbp"   , fmt.Sprintf("%#016x -> %#016x", self.rbp   , rf.rbp   ), false}) }
    if self.rsp    != rf.rsp    { v = append(v, _Reg{"rsp"   , fmt.Sprintf("%#016x -> %#016x", self.rsp   , rf.rsp   ), false}) }
    if self.r8     != rf.r8     { v = append(v, _Reg{"r8"    , fmt.Sprintf("%#016x -> %#016x", self.r8    , rf.r8    ), false}) }
    if self.r9     != rf.r9     { v = append(v, _Reg{"r9"    , fmt.Sprintf("%#016x -> %#016x", self.r9    , rf.r9    ), false}) }
    if self.r10    != rf.r10    { v = append(v, _Reg{"r10"   , fmt.Sprintf("%#016x -> %#016x", self.r10   , rf.r10   ), false}) }
    if self.r11    != rf.r11    { v = append(v, _Reg{"r11"   , fmt.Sprintf("%#016x -> %#016x", self.r11   , rf.r11   ), false}) }
    if self.r12    != rf.r12    { v = append(v, _Reg{"r12"   , fmt.Sprintf("%#016x -> %#016x", self.r12   , rf.r12   ), false}) }
    if self.r13    != rf.r13    { v = append(v, _Reg{"r13"   , fmt.Sprintf("%#016x -> %#016x", self.r13   , rf.r13   ), false}) }
    if self.r14    != rf.r14    { v = append(v, _Reg{"r14"   , fmt.Sprintf("%#016x -> %#016x", self.r14   , rf.r14   ), false}) }
    if self.r15    != rf.r15    { v = append(v, _Reg{"r15"   , fmt.Sprintf("%#016x -> %#016x", self.r15   , rf.r15   ), false}) }
    if self.rflags != rf.rflags { v = append(v, _Reg{"rflags", fmt.Sprintf("%#016x -> %#016x", self.rflags, rf.rflags), false}) }
    if self.cs     != rf.cs     { v = append(v, _Reg{"cs"    , fmt.Sprintf("%#04x -> %#04x"  , self.cs    , rf.cs    ), false}) }
    if self.fs     != rf.fs     { v = append(v, _Reg{"fs"    , fmt.Sprintf("%#04x -> %#04x"  , self.fs    , rf.fs    ), false}) }
    if self.gs     != rf.gs     { v = append(v, _Reg{"gs"    , fmt.Sprintf("%#04x -> %#04x"  , self.gs    , rf.gs    ), false}) }

    /* compare vector registers */
    self.compareXmmStd(&v, rf, indent)
    self.compareXmmExt(&v, rf, indent)
    self.compareYmmStd(&v, rf, indent)
    self.compareYmmExt(&v, rf, indent)
    self.compareZmmAll(&v, rf, indent)
    self.compareMaskQW(&v, rf)
    return v
}

func (self *_RegFileAMD64) format(indent int) (v []_Reg) {
    v = []_Reg {
        { "rax"   , fmt.Sprintf("%#016x %s", self.rax, symbol(self.rax)), false },
        { "rbx"   , fmt.Sprintf("%#016x %s", self.rbx, symbol(self.rbx)), false },
        { "rcx"   , fmt.Sprintf("%#016x %s", self.rcx, symbol(self.rcx)), false },
        { "rdx"   , fmt.Sprintf("%#016x %s", self.rdx, symbol(self.rdx)), false },
        { "rdi"   , fmt.Sprintf("%#016x %s", self.rdi, symbol(self.rdi)), false },
        { "rsi"   , fmt.Sprintf("%#016x %s", self.rsi, symbol(self.rsi)), false },
        { "rbp"   , fmt.Sprintf("%#016x %s", self.rbp, symbol(self.rbp)), false },
        { "rsp"   , fmt.Sprintf("%#016x %s", self.rsp, symbol(self.rsp)), false },
        { "r8"    , fmt.Sprintf("%#016x %s", self.r8 , symbol(self.r8 )), false },
        { "r9"    , fmt.Sprintf("%#016x %s", self.r9 , symbol(self.r9 )), false },
        { "r10"   , fmt.Sprintf("%#016x %s", self.r10, symbol(self.r10)), false },
        { "r11"   , fmt.Sprintf("%#016x %s", self.r11, symbol(self.r11)), false },
        { "r12"   , fmt.Sprintf("%#016x %s", self.r12, symbol(self.r12)), false },
        { "r13"   , fmt.Sprintf("%#016x %s", self.r13, symbol(self.r13)), false },
        { "r14"   , fmt.Sprintf("%#016x %s", self.r14, symbol(self.r14)), false },
        { "r15"   , fmt.Sprintf("%#016x %s", self.r15, symbol(self.r15)), false },
        { "rip"   , fmt.Sprintf("%#016x %s", self.rip, symbol(self.rip)), false },
        { "rflags", fmt.Sprintf("%#016x"   , self.rflags)               , false },
        { "cs"    , fmt.Sprintf("%#04x"    , self.cs)                   , false },
        { "fs"    , fmt.Sprintf("%#04x"    , self.fs)                   , false },
        { "gs"    , fmt.Sprintf("%#04x"    , self.gs)                   , false },
    }

    /* add the vector registers */
    self.formatXmmStd(&v, indent)
    self.formatXmmExt(&v, indent)
    self.formatYmmStd(&v, indent)
    self.formatYmmExt(&v, indent)
    self.formatZmmAll(&v, indent)
    self.formatMaskQW(&v)
    return
}

func (self *_RegFileAMD64) formatXmmStd(v *[]_Reg, indent int) {
    for i := 0; i < 16; i++ {
        *v = append(*v, _Reg {
            reg: fmt.Sprintf("xmm%d", i),
            val: vector(self.xmm[i][:], indent),
            vec: true,
        })
    }
}

func (self *_RegFileAMD64) formatXmmExt(v *[]_Reg, indent int) {
    if hasAVX512VL {
        for i := 16; i < 32; i++ {
            *v = append(*v, _Reg {
                reg: fmt.Sprintf("xmm%d", i),
                val: vector(self.xmm[i][:], indent),
                vec: true,
            })
        }
    }
}

func (self *_RegFileAMD64) formatYmmStd(v *[]_Reg, indent int) {
    if hasAVX {
        for i := 0; i < 16; i++ {
            *v = append(*v, _Reg {
                reg: fmt.Sprintf("ymm%d", i),
                val: vector(self.ymm[i][:], indent),
                vec: true,
            })
        }
    }
}

func (self *_RegFileAMD64) formatYmmExt(v *[]_Reg, indent int) {
    if hasAVX512VL {
        for i := 16; i < 32; i++ {
            *v = append(*v, _Reg {
                reg: fmt.Sprintf("ymm%d", i),
                val: vector(self.ymm[i][:], indent),
                vec: true,
            })
        }
    }
}

func (self *_RegFileAMD64) formatZmmAll(v *[]_Reg, indent int) {
    if hasAVX512F {
        for i := 0; i < 32; i++ {
            *v = append(*v, _Reg {
                reg: fmt.Sprintf("zmm%d", i),
                val: vector(self.zmm[i][:], indent),
                vec: true,
            })
        }
    }
}

func (self *_RegFileAMD64) formatMaskQW(v *[]_Reg) {
    if hasAVX512BW {
        for i := 0; i < 8; i++ {
            *v = append(*v, _Reg {
                reg: fmt.Sprintf("k%d", i),
                val: fmt.Sprintf("%#016x", self.k[i]),
                vec: true,
            })
        }
    } else if hasAVX512F {
        for i := 0; i < 8; i++ {
            *v = append(*v, _Reg {
                reg: fmt.Sprintf("k%d", i),
                val: fmt.Sprintf("%#04x", self.k[i]),
                vec: true,
            })
        }
    }
}

func (self *_RegFileAMD64) compareXmmStd(v *[]_Reg, rf *_RegFileAMD64, indent int) {
    for i := 0; i < 16; i++ {
        if self.xmm[i] != rf.xmm[i] {
            *v = append(*v, _Reg {
                reg: fmt.Sprintf("xmm%d", i),
                val: vecdiff(self.xmm[i][:], rf.xmm[i][:], indent),
                vec: true,
            })
        }
    }
}

func (self *_RegFileAMD64) compareXmmExt(v *[]_Reg, rf *_RegFileAMD64, indent int) {
    if hasAVX512VL {
        for i := 16; i < 32; i++ {
            if self.xmm[i] != rf.xmm[i] {
                *v = append(*v, _Reg {
                    reg: fmt.Sprintf("xmm%d", i),
                    val: vecdiff(self.xmm[i][:], rf.xmm[i][:], indent),
                    vec: true,
                })
            }
        }
    }
}

func (self *_RegFileAMD64) compareYmmStd(v *[]_Reg, rf *_RegFileAMD64, indent int) {
    if hasAVX {
        for i := 0; i < 16; i++ {
            if self.ymm[i] != rf.ymm[i] {
                *v = append(*v, _Reg {
                    reg: fmt.Sprintf("ymm%d", i),
                    val: vecdiff(self.ymm[i][:], rf.ymm[i][:], indent),
                    vec: true,
                })
            }
        }
    }
}

func (self *_RegFileAMD64) compareYmmExt(v *[]_Reg, rf *_RegFileAMD64, indent int) {
    if hasAVX512VL {
        for i := 16; i < 32; i++ {
            if self.ymm[i] != rf.ymm[i] {
                *v = append(*v, _Reg {
                    reg: fmt.Sprintf("ymm%d", i),
                    val: vecdiff(self.ymm[i][:], rf.ymm[i][:], indent),
                    vec: true,
                })
            }
        }
    }
}

func (self *_RegFileAMD64) compareZmmAll(v *[]_Reg, rf *_RegFileAMD64, indent int) {
    if hasAVX512F {
        for i := 0; i < 32; i++ {
            if self.zmm[i] != rf.zmm[i] {
                *v = append(*v, _Reg {
                    reg: fmt.Sprintf("zmm%d", i),
                    val: vecdiff(self.zmm[i][:], rf.zmm[i][:], indent),
                    vec: true,
                })
            }
        }
    }
}

func (self *_RegFileAMD64) compareMaskQW(v *[]_Reg, rf *_RegFileAMD64) {
    if hasAVX512BW {
        for i := 0; i < 8; i++ {
            if self.k[i] != rf.k[i] {
                *v = append(*v, _Reg {
                    reg: fmt.Sprintf("k%d", i),
                    val: fmt.Sprintf("%#016x -> %#016x", self.k[i], rf.k[i]),
                    vec: true,
                })
            }
        }
    } else if hasAVX512F {
        for i := 0; i < 8; i++ {
            if self.k[i] != rf.k[i] {
                *v = append(*v, _Reg {
                    reg: fmt.Sprintf("k%d", i),
                    val: fmt.Sprintf("%#04x -> %#04x", self.k[i], rf.k[i]),
                    vec: true,
                })
            }
        }
    }
}

func init() {
    _regs = new(_RegFileAMD64)
}
