package x86_64

import (
    `fmt`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// ISA represents an extension to x86-64 instruction set.
type ISA uint8

const (
    ISA_CPUID ISA = iota
    ISA_RDTSC
    ISA_RDTSCP
    ISA_CMOV
    ISA_MOVBE
    ISA_POPCNT
    ISA_LZCNT
    ISA_TBM
    ISA_BMI
    ISA_BMI2
    ISA_ADX
    ISA_MMX
    ISA_MMX_PLUS
    ISA_FEMMS
    ISA_3DNOW
    ISA_3DNOW_PLUS
    ISA_SSE
    ISA_SSE2
    ISA_SSE3
    ISA_SSSE3
    ISA_SSE4A
    ISA_SSE4_1
    ISA_SSE4_2
    ISA_FMA3
    ISA_FMA4
    ISA_XOP
    ISA_F16C
    ISA_AVX
    ISA_AVX2
    ISA_AVX512F
    ISA_AVX512BW
    ISA_AVX512DQ
    ISA_AVX512VL
    ISA_AVX512PF
    ISA_AVX512ER
    ISA_AVX512CD
    ISA_AVX512VBMI
    ISA_AVX512IFMA
    ISA_AVX512VPOPCNTDQ
    ISA_AVX512_4VNNIW
    ISA_AVX512_4FMAPS
    ISA_PREFETCH
    ISA_PREFETCHW
    ISA_PREFETCHWT1
    ISA_CLFLUSH
    ISA_CLFLUSHOPT
    ISA_CLWB
    ISA_CLZERO
    ISA_RDRAND
    ISA_RDSEED
    ISA_PCLMULQDQ
    ISA_AES
    ISA_SHA
    ISA_MONITOR
    ISA_MONITORX
    _ISA_MAX = ISA_MONITORX
)

var _ISA_NAMES = [...]string {
    ISA_CPUID           : "CPUID",
    ISA_RDTSC           : "RDTSC",
    ISA_RDTSCP          : "RDTSCP",
    ISA_CMOV            : "CMOV",
    ISA_MOVBE           : "MOVBE",
    ISA_POPCNT          : "POPCNT",
    ISA_LZCNT           : "LZCNT",
    ISA_TBM             : "TBM",
    ISA_BMI             : "BMI",
    ISA_BMI2            : "BMI2",
    ISA_ADX             : "ADX",
    ISA_MMX             : "MMX",
    ISA_MMX_PLUS        : "MMX+",
    ISA_FEMMS           : "FEMMS",
    ISA_3DNOW           : "3dnow!",
    ISA_3DNOW_PLUS      : "3dnow!+",
    ISA_SSE             : "SSE",
    ISA_SSE2            : "SSE2",
    ISA_SSE3            : "SSE3",
    ISA_SSSE3           : "SSSE3",
    ISA_SSE4A           : "SSE4A",
    ISA_SSE4_1          : "SSE4.1",
    ISA_SSE4_2          : "SSE4.2",
    ISA_FMA3            : "FMA3",
    ISA_FMA4            : "FMA4",
    ISA_XOP             : "XOP",
    ISA_F16C            : "F16C",
    ISA_AVX             : "AVX",
    ISA_AVX2            : "AVX2",
    ISA_AVX512F         : "AVX512F",
    ISA_AVX512BW        : "AVX512BW",
    ISA_AVX512DQ        : "AVX512DQ",
    ISA_AVX512VL        : "AVX512VL",
    ISA_AVX512PF        : "AVX512PF",
    ISA_AVX512ER        : "AVX512ER",
    ISA_AVX512CD        : "AVX512CD",
    ISA_AVX512VBMI      : "AVX512VBMI",
    ISA_AVX512IFMA      : "AVX512IFMA",
    ISA_AVX512VPOPCNTDQ : "AVX512VPOPCNTDQ",
    ISA_AVX512_4VNNIW   : "AVX512_4VNNIW",
    ISA_AVX512_4FMAPS   : "AVX512_4FMAPS",
    ISA_PREFETCH        : "PREFETCH",
    ISA_PREFETCHW       : "PREFETCHW",
    ISA_PREFETCHWT1     : "PREFETCHWT1",
    ISA_CLFLUSH         : "CLFLUSH",
    ISA_CLFLUSHOPT      : "CLFLUSHOPT",
    ISA_CLWB            : "CLWB",
    ISA_CLZERO          : "CLZERO",
    ISA_RDRAND          : "RDRAND",
    ISA_RDSEED          : "RDSEED",
    ISA_PCLMULQDQ       : "PCLMULQDQ",
    ISA_AES             : "AES",
    ISA_SHA             : "SHA",
    ISA_MONITOR         : "MONITOR",
    ISA_MONITORX        : "MONITORX",
}

var _ISA_MAPPINGS = map[string]ISA {
    "CPUID"           : ISA_CPUID,
    "RDTSC"           : ISA_RDTSC,
    "RDTSCP"          : ISA_RDTSCP,
    "CMOV"            : ISA_CMOV,
    "MOVBE"           : ISA_MOVBE,
    "POPCNT"          : ISA_POPCNT,
    "LZCNT"           : ISA_LZCNT,
    "TBM"             : ISA_TBM,
    "BMI"             : ISA_BMI,
    "BMI2"            : ISA_BMI2,
    "ADX"             : ISA_ADX,
    "MMX"             : ISA_MMX,
    "MMX+"            : ISA_MMX_PLUS,
    "FEMMS"           : ISA_FEMMS,
    "3dnow!"          : ISA_3DNOW,
    "3dnow!+"         : ISA_3DNOW_PLUS,
    "SSE"             : ISA_SSE,
    "SSE2"            : ISA_SSE2,
    "SSE3"            : ISA_SSE3,
    "SSSE3"           : ISA_SSSE3,
    "SSE4A"           : ISA_SSE4A,
    "SSE4.1"          : ISA_SSE4_1,
    "SSE4.2"          : ISA_SSE4_2,
    "FMA3"            : ISA_FMA3,
    "FMA4"            : ISA_FMA4,
    "XOP"             : ISA_XOP,
    "F16C"            : ISA_F16C,
    "AVX"             : ISA_AVX,
    "AVX2"            : ISA_AVX2,
    "AVX512F"         : ISA_AVX512F,
    "AVX512BW"        : ISA_AVX512BW,
    "AVX512DQ"        : ISA_AVX512DQ,
    "AVX512VL"        : ISA_AVX512VL,
    "AVX512PF"        : ISA_AVX512PF,
    "AVX512ER"        : ISA_AVX512ER,
    "AVX512CD"        : ISA_AVX512CD,
    "AVX512VBMI"      : ISA_AVX512VBMI,
    "AVX512IFMA"      : ISA_AVX512IFMA,
    "AVX512VPOPCNTDQ" : ISA_AVX512VPOPCNTDQ,
    "AVX512_4VNNIW"   : ISA_AVX512_4VNNIW,
    "AVX512_4FMAPS"   : ISA_AVX512_4FMAPS,
    "PREFETCH"        : ISA_PREFETCH,
    "PREFETCHW"       : ISA_PREFETCHW,
    "PREFETCHWT1"     : ISA_PREFETCHWT1,
    "CLFLUSH"         : ISA_CLFLUSH,
    "CLFLUSHOPT"      : ISA_CLFLUSHOPT,
    "CLWB"            : ISA_CLWB,
    "CLZERO"          : ISA_CLZERO,
    "RDRAND"          : ISA_RDRAND,
    "RDSEED"          : ISA_RDSEED,
    "PCLMULQDQ"       : ISA_PCLMULQDQ,
    "AES"             : ISA_AES,
    "SHA"             : ISA_SHA,
    "MONITOR"         : ISA_MONITOR,
    "MONITORX"        : ISA_MONITORX,
}

func (self ISA) Sealed(tag.Tag) {}
func (self ISA) FeatureID() uint8 { return uint8(self) }

func (self ISA) String() string {
    if self <= _ISA_MAX {
        return _ISA_NAMES[self]
    } else {
        return fmt.Sprintf("(invalid: %#x)", uint64(self))
    }
}

// ParseISA parses name into ISA, it will panic if the name is invalid.
func ParseISA(name string) ISA {
    if v, ok := _ISA_MAPPINGS[name]; ok {
        return v
    } else {
        panic("invalid ISA name: " + name)
    }
}

type (
	_ProgramFactory struct{}
)

func (_ProgramFactory) Sealed(tag.Tag) {}
func (_ProgramFactory) CreateProgram() asm.Program { return newProgram(x86_64) }

var x86_64 = asm.RegisterArch(
    "x86_64",
    _ISA_MAX,
    new(_ProgramFactory),
)
