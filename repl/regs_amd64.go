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
func dumpregs(_ *_RegFileAMD64)

func (self *_RegFileAMD64) Dump(indent int) []_Reg {
    dumpregs(self)
    return self.format(indent)
}

func (self *_RegFileAMD64) format(indent int) (v []_Reg) {
    v = []_Reg {
        { "rax"    , fmt.Sprintf("%#016x", self.rax)    , false },
        { "rbx"    , fmt.Sprintf("%#016x", self.rbx)    , false },
        { "rcx"    , fmt.Sprintf("%#016x", self.rcx)    , false },
        { "rdx"    , fmt.Sprintf("%#016x", self.rdx)    , false },
        { "rdi"    , fmt.Sprintf("%#016x", self.rdi)    , false },
        { "rsi"    , fmt.Sprintf("%#016x", self.rsi)    , false },
        { "rbp"    , fmt.Sprintf("%#016x", self.rbp)    , false },
        { "rsp"    , fmt.Sprintf("%#016x", self.rsp)    , false },
        { "r8"     , fmt.Sprintf("%#016x", self.r8)     , false },
        { "r9"     , fmt.Sprintf("%#016x", self.r9)     , false },
        { "r10"    , fmt.Sprintf("%#016x", self.r10)    , false },
        { "r11"    , fmt.Sprintf("%#016x", self.r11)    , false },
        { "r12"    , fmt.Sprintf("%#016x", self.r12)    , false },
        { "r13"    , fmt.Sprintf("%#016x", self.r13)    , false },
        { "r14"    , fmt.Sprintf("%#016x", self.r14)    , false },
        { "r15"    , fmt.Sprintf("%#016x", self.r15)    , false },
        { "rip"    , fmt.Sprintf("%#016x", self.rip)    , false },
        { "rflags" , fmt.Sprintf("%#016x", self.rflags) , false },
        { "cs"     , fmt.Sprintf("%#04x", self.cs)      , false },
        { "fs"     , fmt.Sprintf("%#04x", self.fs)      , false },
        { "gs"     , fmt.Sprintf("%#04x", self.gs)      , false },
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

func init() {
    _regs = new(_RegFileAMD64)
}
