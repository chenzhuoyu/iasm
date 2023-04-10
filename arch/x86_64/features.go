package x86_64

import (
    `fmt`
)

// Feature represents an extension to x86-64 CPU feature.
type Feature uint8

const (
    FEAT_CPUID Feature = iota
    FEAT_RDTSC
    FEAT_RDTSCP
    FEAT_CMOV
    FEAT_MOVBE
    FEAT_POPCNT
    FEAT_LZCNT
    FEAT_TBM
    FEAT_BMI
    FEAT_BMI2
    FEAT_ADX
    FEAT_MMX
    FEAT_MMX_PLUS
    FEAT_FEMMS
    FEAT_3DNOW
    FEAT_3DNOW_PLUS
    FEAT_SSE
    FEAT_SSE2
    FEAT_SSE3
    FEAT_SSSE3
    FEAT_SSE4A
    FEAT_SSE4_1
    FEAT_SSE4_2
    FEAT_FMA3
    FEAT_FMA4
    FEAT_XOP
    FEAT_F16C
    FEAT_AVX
    FEAT_AVX2
    FEAT_AVX512F
    FEAT_AVX512BW
    FEAT_AVX512DQ
    FEAT_AVX512VL
    FEAT_AVX512PF
    FEAT_AVX512ER
    FEAT_AVX512CD
    FEAT_AVX512VBMI
    FEAT_AVX512IFMA
    FEAT_AVX512VPOPCNTDQ
    FEAT_AVX512_4VNNIW
    FEAT_AVX512_4FMAPS
    FEAT_PREFETCH
    FEAT_PREFETCHW
    FEAT_PREFETCHWT1
    FEAT_CLFLUSH
    FEAT_CLFLUSHOPT
    FEAT_CLWB
    FEAT_CLZERO
    FEAT_RDRAND
    FEAT_RDSEED
    FEAT_PCLMULQDQ
    FEAT_AES
    FEAT_SHA
    FEAT_MONITOR
    FEAT_MONITORX
    _FEAT_MAX = FEAT_MONITORX
)

var _FEAT_NAMES = [...]string {
    FEAT_CPUID           : "CPUID",
    FEAT_RDTSC           : "RDTSC",
    FEAT_RDTSCP          : "RDTSCP",
    FEAT_CMOV            : "CMOV",
    FEAT_MOVBE           : "MOVBE",
    FEAT_POPCNT          : "POPCNT",
    FEAT_LZCNT           : "LZCNT",
    FEAT_TBM             : "TBM",
    FEAT_BMI             : "BMI",
    FEAT_BMI2            : "BMI2",
    FEAT_ADX             : "ADX",
    FEAT_MMX             : "MMX",
    FEAT_MMX_PLUS        : "MMX+",
    FEAT_FEMMS           : "FEMMS",
    FEAT_3DNOW           : "3dnow!",
    FEAT_3DNOW_PLUS      : "3dnow!+",
    FEAT_SSE             : "SSE",
    FEAT_SSE2            : "SSE2",
    FEAT_SSE3            : "SSE3",
    FEAT_SSSE3           : "SSSE3",
    FEAT_SSE4A           : "SSE4A",
    FEAT_SSE4_1          : "SSE4.1",
    FEAT_SSE4_2          : "SSE4.2",
    FEAT_FMA3            : "FMA3",
    FEAT_FMA4            : "FMA4",
    FEAT_XOP             : "XOP",
    FEAT_F16C            : "F16C",
    FEAT_AVX             : "AVX",
    FEAT_AVX2            : "AVX2",
    FEAT_AVX512F         : "AVX512F",
    FEAT_AVX512BW        : "AVX512BW",
    FEAT_AVX512DQ        : "AVX512DQ",
    FEAT_AVX512VL        : "AVX512VL",
    FEAT_AVX512PF        : "AVX512PF",
    FEAT_AVX512ER        : "AVX512ER",
    FEAT_AVX512CD        : "AVX512CD",
    FEAT_AVX512VBMI      : "AVX512VBMI",
    FEAT_AVX512IFMA      : "AVX512IFMA",
    FEAT_AVX512VPOPCNTDQ : "AVX512VPOPCNTDQ",
    FEAT_AVX512_4VNNIW   : "AVX512_4VNNIW",
    FEAT_AVX512_4FMAPS   : "AVX512_4FMAPS",
    FEAT_PREFETCH        : "PREFETCH",
    FEAT_PREFETCHW       : "PREFETCHW",
    FEAT_PREFETCHWT1     : "PREFETCHWT1",
    FEAT_CLFLUSH         : "CLFLUSH",
    FEAT_CLFLUSHOPT      : "CLFLUSHOPT",
    FEAT_CLWB            : "CLWB",
    FEAT_CLZERO          : "CLZERO",
    FEAT_RDRAND          : "RDRAND",
    FEAT_RDSEED          : "RDSEED",
    FEAT_PCLMULQDQ       : "PCLMULQDQ",
    FEAT_AES             : "AES",
    FEAT_SHA             : "SHA",
    FEAT_MONITOR         : "MONITOR",
    FEAT_MONITORX        : "MONITORX",
}

var _FEAT_MAPPINGS = map[string]Feature{
    "CPUID"           : FEAT_CPUID,
    "RDTSC"           : FEAT_RDTSC,
    "RDTSCP"          : FEAT_RDTSCP,
    "CMOV"            : FEAT_CMOV,
    "MOVBE"           : FEAT_MOVBE,
    "POPCNT"          : FEAT_POPCNT,
    "LZCNT"           : FEAT_LZCNT,
    "TBM"             : FEAT_TBM,
    "BMI"             : FEAT_BMI,
    "BMI2"            : FEAT_BMI2,
    "ADX"             : FEAT_ADX,
    "MMX"             : FEAT_MMX,
    "MMX+"            : FEAT_MMX_PLUS,
    "FEMMS"           : FEAT_FEMMS,
    "3dnow!"          : FEAT_3DNOW,
    "3dnow!+"         : FEAT_3DNOW_PLUS,
    "SSE"             : FEAT_SSE,
    "SSE2"            : FEAT_SSE2,
    "SSE3"            : FEAT_SSE3,
    "SSSE3"           : FEAT_SSSE3,
    "SSE4A"           : FEAT_SSE4A,
    "SSE4.1"          : FEAT_SSE4_1,
    "SSE4.2"          : FEAT_SSE4_2,
    "FMA3"            : FEAT_FMA3,
    "FMA4"            : FEAT_FMA4,
    "XOP"             : FEAT_XOP,
    "F16C"            : FEAT_F16C,
    "AVX"             : FEAT_AVX,
    "AVX2"            : FEAT_AVX2,
    "AVX512F"         : FEAT_AVX512F,
    "AVX512BW"        : FEAT_AVX512BW,
    "AVX512DQ"        : FEAT_AVX512DQ,
    "AVX512VL"        : FEAT_AVX512VL,
    "AVX512PF"        : FEAT_AVX512PF,
    "AVX512ER"        : FEAT_AVX512ER,
    "AVX512CD"        : FEAT_AVX512CD,
    "AVX512VBMI"      : FEAT_AVX512VBMI,
    "AVX512IFMA"      : FEAT_AVX512IFMA,
    "AVX512VPOPCNTDQ" : FEAT_AVX512VPOPCNTDQ,
    "AVX512_4VNNIW"   : FEAT_AVX512_4VNNIW,
    "AVX512_4FMAPS"   : FEAT_AVX512_4FMAPS,
    "PREFETCH"        : FEAT_PREFETCH,
    "PREFETCHW"       : FEAT_PREFETCHW,
    "PREFETCHWT1"     : FEAT_PREFETCHWT1,
    "CLFLUSH"         : FEAT_CLFLUSH,
    "CLFLUSHOPT"      : FEAT_CLFLUSHOPT,
    "CLWB"            : FEAT_CLWB,
    "CLZERO"          : FEAT_CLZERO,
    "RDRAND"          : FEAT_RDRAND,
    "RDSEED"          : FEAT_RDSEED,
    "PCLMULQDQ"       : FEAT_PCLMULQDQ,
    "AES"             : FEAT_AES,
    "SHA"             : FEAT_SHA,
    "MONITOR"         : FEAT_MONITOR,
    "MONITORX"        : FEAT_MONITORX,
}

func (self Feature) String() string {
    if self <= _FEAT_MAX {
        return _FEAT_NAMES[self]
    } else {
        return fmt.Sprintf("(invalid: %#x)", uint64(self))
    }
}

func (self Feature) FeatureID() uint8 {
    return uint8(self)
}

// ParseFeature parses name into Feature, it will panic if the name is invalid.
func ParseFeature(name string) Feature {
    if v, ok := _FEAT_MAPPINGS[name]; ok {
        return v
    } else {
        panic("invalid feature name: " + name)
    }
}
