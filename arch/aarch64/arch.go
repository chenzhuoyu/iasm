package aarch64

import (
    `fmt`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Feature represents one of the arch-specific features.
type Feature uint8

const (
    FEAT_AES Feature = iota
    FEAT_ASMv8p2
    FEAT_B16B16
    FEAT_BF16
    FEAT_BRBE
    FEAT_BTI
    FEAT_CHK
    FEAT_CLRBHB
    FEAT_CRC32
    FEAT_CSSC
    FEAT_D128
    FEAT_DGH
    FEAT_DotProd
    FEAT_EBEP
    FEAT_F32MM
    FEAT_F64MM
    FEAT_FCMA
    FEAT_FHM
    FEAT_FP16
    FEAT_FRINTTS
    FEAT_FlagM
    FEAT_FlagM2
    FEAT_GCS
    FEAT_HBC
    FEAT_I8MM
    FEAT_ITE
    FEAT_JSCVT
    FEAT_LOR
    FEAT_LRCPC
    FEAT_LRCPC2
    FEAT_LRCPC3
    FEAT_LS64
    FEAT_LS64_ACCDATA
    FEAT_LS64_V
    FEAT_LSE
    FEAT_LSE128
    FEAT_MOPS
    FEAT_MTE
    FEAT_MTE2
    FEAT_PAuth
    FEAT_RAS
    FEAT_RDM
    FEAT_RPRFM
    FEAT_SB
    FEAT_SHA1
    FEAT_SHA256
    FEAT_SHA3
    FEAT_SHA512
    FEAT_SM3
    FEAT_SM4
    FEAT_SME
    FEAT_SME2
    FEAT_SME2p1
    FEAT_SME_F16F16
    FEAT_SME_F64F64
    FEAT_SME_I16I64
    FEAT_SPE
    FEAT_SPECRES
    FEAT_SPECRES2
    FEAT_SVE2p1
    FEAT_SVE_AES
    FEAT_SVE_BitPerm
    FEAT_SVE_PMULL128
    FEAT_SVE_SHA3
    FEAT_SVE_SM4
    FEAT_SYSINSTR128
    FEAT_SYSREG128
    FEAT_THE
    FEAT_TME
    FEAT_TRF
    FEAT_WFxT
    FEAT_XS
    _FEAT_MAX = FEAT_XS
)

var _FEAT_NAMES = [...]string {
    FEAT_AES          : "FEAT_AES",
    FEAT_ASMv8p2      : "FEAT_ASMv8p2",
    FEAT_B16B16       : "FEAT_B16B16",
    FEAT_BF16         : "FEAT_BF16",
    FEAT_BRBE         : "FEAT_BRBE",
    FEAT_BTI          : "FEAT_BTI",
    FEAT_CHK          : "FEAT_CHK",
    FEAT_CLRBHB       : "FEAT_CLRBHB",
    FEAT_CRC32        : "FEAT_CRC32",
    FEAT_CSSC         : "FEAT_CSSC",
    FEAT_D128         : "FEAT_D128",
    FEAT_DGH          : "FEAT_DGH",
    FEAT_DotProd      : "FEAT_DotProd",
    FEAT_EBEP         : "FEAT_EBEP",
    FEAT_F32MM        : "FEAT_F32MM",
    FEAT_F64MM        : "FEAT_F64MM",
    FEAT_FCMA         : "FEAT_FCMA",
    FEAT_FHM          : "FEAT_FHM",
    FEAT_FP16         : "FEAT_FP16",
    FEAT_FRINTTS      : "FEAT_FRINTTS",
    FEAT_FlagM        : "FEAT_FlagM",
    FEAT_FlagM2       : "FEAT_FlagM2",
    FEAT_GCS          : "FEAT_GCS",
    FEAT_HBC          : "FEAT_HBC",
    FEAT_I8MM         : "FEAT_I8MM",
    FEAT_ITE          : "FEAT_ITE",
    FEAT_JSCVT        : "FEAT_JSCVT",
    FEAT_LOR          : "FEAT_LOR",
    FEAT_LRCPC        : "FEAT_LRCPC",
    FEAT_LRCPC2       : "FEAT_LRCPC2",
    FEAT_LRCPC3       : "FEAT_LRCPC3",
    FEAT_LS64         : "FEAT_LS64",
    FEAT_LS64_ACCDATA : "FEAT_LS64_ACCDATA",
    FEAT_LS64_V       : "FEAT_LS64_V",
    FEAT_LSE          : "FEAT_LSE",
    FEAT_LSE128       : "FEAT_LSE128",
    FEAT_MOPS         : "FEAT_MOPS",
    FEAT_MTE          : "FEAT_MTE",
    FEAT_MTE2         : "FEAT_MTE2",
    FEAT_PAuth        : "FEAT_PAuth",
    FEAT_RAS          : "FEAT_RAS",
    FEAT_RDM          : "FEAT_RDM",
    FEAT_RPRFM        : "FEAT_RPRFM",
    FEAT_SB           : "FEAT_SB",
    FEAT_SHA1         : "FEAT_SHA1",
    FEAT_SHA256       : "FEAT_SHA256",
    FEAT_SHA3         : "FEAT_SHA3",
    FEAT_SHA512       : "FEAT_SHA512",
    FEAT_SM3          : "FEAT_SM3",
    FEAT_SM4          : "FEAT_SM4",
    FEAT_SME          : "FEAT_SME",
    FEAT_SME2         : "FEAT_SME2",
    FEAT_SME2p1       : "FEAT_SME2p1",
    FEAT_SME_F16F16   : "FEAT_SME_F16F16",
    FEAT_SME_F64F64   : "FEAT_SME_F64F64",
    FEAT_SME_I16I64   : "FEAT_SME_I16I64",
    FEAT_SPE          : "FEAT_SPE",
    FEAT_SPECRES      : "FEAT_SPECRES",
    FEAT_SPECRES2     : "FEAT_SPECRES2",
    FEAT_SVE2p1       : "FEAT_SVE2p1",
    FEAT_SVE_AES      : "FEAT_SVE_AES",
    FEAT_SVE_BitPerm  : "FEAT_SVE_BitPerm",
    FEAT_SVE_PMULL128 : "FEAT_SVE_PMULL128",
    FEAT_SVE_SHA3     : "FEAT_SVE_SHA3",
    FEAT_SVE_SM4      : "FEAT_SVE_SM4",
    FEAT_SYSINSTR128  : "FEAT_SYSINSTR128",
    FEAT_SYSREG128    : "FEAT_SYSREG128",
    FEAT_THE          : "FEAT_THE",
    FEAT_TME          : "FEAT_TME",
    FEAT_TRF          : "FEAT_TRF",
    FEAT_WFxT         : "FEAT_WFxT",
    FEAT_XS           : "FEAT_XS",
}

var _FEAT_MAPPINGS = map[string]Feature {
    "FEAT_AES"          : FEAT_AES,
    "FEAT_ASMv8p2"      : FEAT_ASMv8p2,
    "FEAT_B16B16"       : FEAT_B16B16,
    "FEAT_BF16"         : FEAT_BF16,
    "FEAT_BRBE"         : FEAT_BRBE,
    "FEAT_BTI"          : FEAT_BTI,
    "FEAT_CHK"          : FEAT_CHK,
    "FEAT_CLRBHB"       : FEAT_CLRBHB,
    "FEAT_CRC32"        : FEAT_CRC32,
    "FEAT_CSSC"         : FEAT_CSSC,
    "FEAT_D128"         : FEAT_D128,
    "FEAT_DGH"          : FEAT_DGH,
    "FEAT_DotProd"      : FEAT_DotProd,
    "FEAT_EBEP"         : FEAT_EBEP,
    "FEAT_F32MM"        : FEAT_F32MM,
    "FEAT_F64MM"        : FEAT_F64MM,
    "FEAT_FCMA"         : FEAT_FCMA,
    "FEAT_FHM"          : FEAT_FHM,
    "FEAT_FP16"         : FEAT_FP16,
    "FEAT_FRINTTS"      : FEAT_FRINTTS,
    "FEAT_FlagM"        : FEAT_FlagM,
    "FEAT_FlagM2"       : FEAT_FlagM2,
    "FEAT_GCS"          : FEAT_GCS,
    "FEAT_HBC"          : FEAT_HBC,
    "FEAT_I8MM"         : FEAT_I8MM,
    "FEAT_ITE"          : FEAT_ITE,
    "FEAT_JSCVT"        : FEAT_JSCVT,
    "FEAT_LOR"          : FEAT_LOR,
    "FEAT_LRCPC"        : FEAT_LRCPC,
    "FEAT_LRCPC2"       : FEAT_LRCPC2,
    "FEAT_LRCPC3"       : FEAT_LRCPC3,
    "FEAT_LS64"         : FEAT_LS64,
    "FEAT_LS64_ACCDATA" : FEAT_LS64_ACCDATA,
    "FEAT_LS64_V"       : FEAT_LS64_V,
    "FEAT_LSE"          : FEAT_LSE,
    "FEAT_LSE128"       : FEAT_LSE128,
    "FEAT_MOPS"         : FEAT_MOPS,
    "FEAT_MTE"          : FEAT_MTE,
    "FEAT_MTE2"         : FEAT_MTE2,
    "FEAT_PAuth"        : FEAT_PAuth,
    "FEAT_RAS"          : FEAT_RAS,
    "FEAT_RDM"          : FEAT_RDM,
    "FEAT_RPRFM"        : FEAT_RPRFM,
    "FEAT_SB"           : FEAT_SB,
    "FEAT_SHA1"         : FEAT_SHA1,
    "FEAT_SHA256"       : FEAT_SHA256,
    "FEAT_SHA3"         : FEAT_SHA3,
    "FEAT_SHA512"       : FEAT_SHA512,
    "FEAT_SM3"          : FEAT_SM3,
    "FEAT_SM4"          : FEAT_SM4,
    "FEAT_SME"          : FEAT_SME,
    "FEAT_SME2"         : FEAT_SME2,
    "FEAT_SME2p1"       : FEAT_SME2p1,
    "FEAT_SME_F16F16"   : FEAT_SME_F16F16,
    "FEAT_SME_F64F64"   : FEAT_SME_F64F64,
    "FEAT_SME_I16I64"   : FEAT_SME_I16I64,
    "FEAT_SPE"          : FEAT_SPE,
    "FEAT_SPECRES"      : FEAT_SPECRES,
    "FEAT_SPECRES2"     : FEAT_SPECRES2,
    "FEAT_SVE2p1"       : FEAT_SVE2p1,
    "FEAT_SVE_AES"      : FEAT_SVE_AES,
    "FEAT_SVE_BitPerm"  : FEAT_SVE_BitPerm,
    "FEAT_SVE_PMULL128" : FEAT_SVE_PMULL128,
    "FEAT_SVE_SHA3"     : FEAT_SVE_SHA3,
    "FEAT_SVE_SM4"      : FEAT_SVE_SM4,
    "FEAT_SYSINSTR128"  : FEAT_SYSINSTR128,
    "FEAT_SYSREG128"    : FEAT_SYSREG128,
    "FEAT_THE"          : FEAT_THE,
    "FEAT_TME"          : FEAT_TME,
    "FEAT_TRF"          : FEAT_TRF,
    "FEAT_WFxT"         : FEAT_WFxT,
    "FEAT_XS"           : FEAT_XS,
}

func (self Feature) Sealed(tag.Tag)   {}
func (self Feature) FeatureID() uint8 { return uint8(self) }

func (self Feature) String() string {
    if self <= _FEAT_MAX {
        return _FEAT_NAMES[self]
    } else {
        return fmt.Sprintf("(invalid: %#x)", uint64(self))
    }
}

// ParseFeature parses name into Feature, it will panic if the name is invalid.
func ParseFeature(name string) Feature {
    if v, ok := _FEAT_MAPPINGS[name]; ok {
        return v
    } else {
        panic("invalid feature name: " + name)
    }
}

type (
	_ProgramFactory struct{}
)

func (_ProgramFactory) Sealed(tag.Tag) {}
func (_ProgramFactory) CreateProgram() asm.Program { return newProgram(aarch64) }

var aarch64 = asm.RegisterArch(
    "aarch64",
    _FEAT_MAX,
    new(_ProgramFactory),
)
