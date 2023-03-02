package aarch64

import (
    `fmt`

    `github.com/chenzhuoyu/iasm/internal/rt`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// SIMDVectorStructure represents the structure of a V register, specifically used
// in the single-structure version of {LD,ST}{1,2,3,4} instructions.
type SIMDVectorStructure uint8

const (
    VecB SIMDVectorStructure = iota
    VecH
    VecS
    VecD
)

func (self SIMDVectorStructure) String() string {
    switch self {
        case VecB : return "B"
        case VecH : return "H"
        case VecS : return "S"
        case VecD : return "D"
        default   : panic("aarch64: invalid SIMD vector structure")
    }
}

func (self SIMDVectorStructure) MaxIndex() uint8 {
    switch self {
        case VecB : return 15
        case VecH : return 7
        case VecS : return 3
        case VecD : return 1
        default   : panic("aarch64: invalid SIMD vector structure")
    }
}

// SIMDVectorArrangement represents the data arrangement of a V register.
type SIMDVectorArrangement uint8

const (
    Vec8B SIMDVectorArrangement = iota
    Vec16B
    Vec4H
    Vec8H
    Vec2S
    Vec4S
    Vec1D
    Vec2D
)

func (self SIMDVectorArrangement) String() string {
    switch self {
        case Vec8B  : return "8B"
        case Vec16B : return "16B"
        case Vec4H  : return "4H"
        case Vec8H  : return "8H"
        case Vec2S  : return "2S"
        case Vec4S  : return "4S"
        case Vec1D  : return "1D"
        case Vec2D  : return "2D"
        default     : panic("aarch64: invalid SIMD vector arrangement")
    }
}

// SIMDRegister8 represents a vector with 8-bit elements (B0 ~ B31).
type SIMDRegister8 uint8

// SIMDRegister16 represents a vector with 16 bit elements (H0 ~ H31).
type SIMDRegister16 uint8

// SIMDRegister32 represents a vector with 32-bit elements (S0 ~ S31).
type SIMDRegister32 uint8

// SIMDRegister64 represents a vector with 64-bit elements (D0 ~ D31).
type SIMDRegister64 uint8

// SIMDRegister128 represents a single 128-bit value (Q0 ~ Q31).
type SIMDRegister128 uint8

// SIMDRegister128r represents an indexed vector base (V{0~31}.{B,H,S,D}).
type SIMDRegister128r uint8

// SIMDRegister128v represents a SIMD vector with certain arrangement (V{0~31}.{8B,16B,4H,8H,2S,4S,1D,2D}).
type SIMDRegister128v uint8

// SIMDRegister128i represents an indexed SIMD vector (V{0~31}.{B,H,S,D}[<index>]).
type SIMDRegister128i interface {
    tag.Sealed
    fmt.Stringer
    ID() uint8
    Index() uint8
    Structure() SIMDVectorStructure
    IndexedRegister()
}

func (SIMDRegister8)    Sealed(tag.Tag) {}
func (SIMDRegister16)   Sealed(tag.Tag) {}
func (SIMDRegister32)   Sealed(tag.Tag) {}
func (SIMDRegister64)   Sealed(tag.Tag) {}
func (SIMDRegister128)  Sealed(tag.Tag) {}
func (SIMDRegister128v) Sealed(tag.Tag) {}

func (self SIMDRegister8)    ID() uint8 { return uint8(self) }
func (self SIMDRegister16)   ID() uint8 { return uint8(self) }
func (self SIMDRegister32)   ID() uint8 { return uint8(self) }
func (self SIMDRegister64)   ID() uint8 { return uint8(self) }
func (self SIMDRegister128)  ID() uint8 { return uint8(self) }
func (self SIMDRegister128r) ID() uint8 { return uint8(self & 0x1f) }
func (self SIMDRegister128v) ID() uint8 { return uint8(self & 0x1f) }

func (self SIMDRegister8)    String() string { return fmt.Sprintf("b%d", self.ID()) }
func (self SIMDRegister16)   String() string { return fmt.Sprintf("h%d", self.ID()) }
func (self SIMDRegister32)   String() string { return fmt.Sprintf("s%d", self.ID()) }
func (self SIMDRegister64)   String() string { return fmt.Sprintf("d%d", self.ID()) }
func (self SIMDRegister128)  String() string { return fmt.Sprintf("q%d", self.ID()) }
func (self SIMDRegister128r) String() string { return fmt.Sprintf("v%d.%s", self.ID(), self.Structure()) }
func (self SIMDRegister128v) String() string { return fmt.Sprintf("v%d.%s", self.ID(), self.Arrangement()) }

func (self SIMDRegister128r) Structure()   SIMDVectorStructure   { return SIMDVectorStructure(self >> 5) }
func (self SIMDRegister128v) Arrangement() SIMDVectorArrangement { return SIMDVectorArrangement(self >> 5) }

type (
	_Indexed128r uint16
)

var (
    _valIndexed128r  = _Indexed128r(0)
    _typeIndexed128r = rt.TypeOf(_valIndexed128r)
)

func (_Indexed128r) Sealed(tag.Tag)   {}
func (_Indexed128r) IndexedRegister() {}

func (self _Indexed128r) ID() uint8 {
    return uint8(self & 0x1f)
}

func (self _Indexed128r) Index() uint8 {
    return uint8(self >> 8)
}

func (self _Indexed128r) String() string {
    return fmt.Sprintf("v%d.%s[%d]", self.ID(), self.Structure(), self.Index())
}

func (self _Indexed128r) Structure() SIMDVectorStructure {
    return SIMDVectorStructure(uint8(self) >> 5)
}

func (self SIMDRegister128r) Index(i uint8) SIMDRegister128i {
    return self.withIndex(_typeIndexed128r, i).(SIMDRegister128i)
}

func (self SIMDRegister128r) withIndex(ty *rt.GoType, i uint8) interface{} {
    if i > self.Structure().MaxIndex() {
        panic("aarc64: vector register index out of bounds")
    } else {
        return rt.WrapUint16(ty, (uint16(i) << 8) | uint16(self))
    }
}

const (
    B0 SIMDRegister8 = iota
    B1
    B2
    B3
    B4
    B5
    B6
    B7
    B8
    B9
    B10
    B11
    B12
    B13
    B14
    B15
    B16
    B17
    B18
    B19
    B20
    B21
    B22
    B23
    B24
    B25
    B26
    B27
    B28
    B29
    B30
    B31
)

const (
    H0 SIMDRegister16 = iota
    H1
    H2
    H3
    H4
    H5
    H6
    H7
    H8
    H9
    H10
    H11
    H12
    H13
    H14
    H15
    H16
    H17
    H18
    H19
    H20
    H21
    H22
    H23
    H24
    H25
    H26
    H27
    H28
    H29
    H30
    H31
)

const (
    S0 SIMDRegister32 = iota
    S1
    S2
    S3
    S4
    S5
    S6
    S7
    S8
    S9
    S10
    S11
    S12
    S13
    S14
    S15
    S16
    S17
    S18
    S19
    S20
    S21
    S22
    S23
    S24
    S25
    S26
    S27
    S28
    S29
    S30
    S31
)

const (
    D0 SIMDRegister64 = iota
    D1
    D2
    D3
    D4
    D5
    D6
    D7
    D8
    D9
    D10
    D11
    D12
    D13
    D14
    D15
    D16
    D17
    D18
    D19
    D20
    D21
    D22
    D23
    D24
    D25
    D26
    D27
    D28
    D29
    D30
    D31
)

const (
    Q0 SIMDRegister128 = iota
    Q1
    Q2
    Q3
    Q4
    Q5
    Q6
    Q7
    Q8
    Q9
    Q10
    Q11
    Q12
    Q13
    Q14
    Q15
    Q16
    Q17
    Q18
    Q19
    Q20
    Q21
    Q22
    Q23
    Q24
    Q25
    Q26
    Q27
    Q28
    Q29
    Q30
    Q31
)

const (
    V0_B = SIMDRegister128r(VecB << 5) | iota
    V1_B
    V2_B
    V3_B
    V4_B
    V5_B
    V6_B
    V7_B
    V8_B
    V9_B
    V10_B
    V11_B
    V12_B
    V13_B
    V14_B
    V15_B
    V16_B
    V17_B
    V18_B
    V19_B
    V20_B
    V21_B
    V22_B
    V23_B
    V24_B
    V25_B
    V26_B
    V27_B
    V28_B
    V29_B
    V30_B
    V31_B
)

const (
    V0_H = SIMDRegister128r(VecH << 5) | iota
    V1_H
    V2_H
    V3_H
    V4_H
    V5_H
    V6_H
    V7_H
    V8_H
    V9_H
    V10_H
    V11_H
    V12_H
    V13_H
    V14_H
    V15_H
    V16_H
    V17_H
    V18_H
    V19_H
    V20_H
    V21_H
    V22_H
    V23_H
    V24_H
    V25_H
    V26_H
    V27_H
    V28_H
    V29_H
    V30_H
    V31_H
)

const (
    V0_S = SIMDRegister128r(VecS << 5) | iota
    V1_S
    V2_S
    V3_S
    V4_S
    V5_S
    V6_S
    V7_S
    V8_S
    V9_S
    V10_S
    V11_S
    V12_S
    V13_S
    V14_S
    V15_S
    V16_S
    V17_S
    V18_S
    V19_S
    V20_S
    V21_S
    V22_S
    V23_S
    V24_S
    V25_S
    V26_S
    V27_S
    V28_S
    V29_S
    V30_S
    V31_S
)

const (
    V0_D = SIMDRegister128r(VecD << 5) | iota
    V1_D
    V2_D
    V3_D
    V4_D
    V5_D
    V6_D
    V7_D
    V8_D
    V9_D
    V10_D
    V11_D
    V12_D
    V13_D
    V14_D
    V15_D
    V16_D
    V17_D
    V18_D
    V19_D
    V20_D
    V21_D
    V22_D
    V23_D
    V24_D
    V25_D
    V26_D
    V27_D
    V28_D
    V29_D
    V30_D
    V31_D
)

const (
    V0_8B = SIMDRegister128v(Vec8B << 5) | iota
    V1_8B
    V2_8B
    V3_8B
    V4_8B
    V5_8B
    V6_8B
    V7_8B
    V8_8B
    V9_8B
    V10_8B
    V11_8B
    V12_8B
    V13_8B
    V14_8B
    V15_8B
    V16_8B
    V17_8B
    V18_8B
    V19_8B
    V20_8B
    V21_8B
    V22_8B
    V23_8B
    V24_8B
    V25_8B
    V26_8B
    V27_8B
    V28_8B
    V29_8B
    V30_8B
    V31_8B
)

const (
    V0_16B = SIMDRegister128v(Vec16B << 5) | iota
    V1_16B
    V2_16B
    V3_16B
    V4_16B
    V5_16B
    V6_16B
    V7_16B
    V8_16B
    V9_16B
    V10_16B
    V11_16B
    V12_16B
    V13_16B
    V14_16B
    V15_16B
    V16_16B
    V17_16B
    V18_16B
    V19_16B
    V20_16B
    V21_16B
    V22_16B
    V23_16B
    V24_16B
    V25_16B
    V26_16B
    V27_16B
    V28_16B
    V29_16B
    V30_16B
    V31_16B
)

const (
    V0_4H = SIMDRegister128v(Vec4H << 5) | iota
    V1_4H
    V2_4H
    V3_4H
    V4_4H
    V5_4H
    V6_4H
    V7_4H
    V8_4H
    V9_4H
    V10_4H
    V11_4H
    V12_4H
    V13_4H
    V14_4H
    V15_4H
    V16_4H
    V17_4H
    V18_4H
    V19_4H
    V20_4H
    V21_4H
    V22_4H
    V23_4H
    V24_4H
    V25_4H
    V26_4H
    V27_4H
    V28_4H
    V29_4H
    V30_4H
    V31_4H
)

const (
    V0_8H = SIMDRegister128v(Vec8H << 5) | iota
    V1_8H
    V2_8H
    V3_8H
    V4_8H
    V5_8H
    V6_8H
    V7_8H
    V8_8H
    V9_8H
    V10_8H
    V11_8H
    V12_8H
    V13_8H
    V14_8H
    V15_8H
    V16_8H
    V17_8H
    V18_8H
    V19_8H
    V20_8H
    V21_8H
    V22_8H
    V23_8H
    V24_8H
    V25_8H
    V26_8H
    V27_8H
    V28_8H
    V29_8H
    V30_8H
    V31_8H
)

const (
    V0_2S = SIMDRegister128v(Vec2S << 5) | iota
    V1_2S
    V2_2S
    V3_2S
    V4_2S
    V5_2S
    V6_2S
    V7_2S
    V8_2S
    V9_2S
    V10_2S
    V11_2S
    V12_2S
    V13_2S
    V14_2S
    V15_2S
    V16_2S
    V17_2S
    V18_2S
    V19_2S
    V20_2S
    V21_2S
    V22_2S
    V23_2S
    V24_2S
    V25_2S
    V26_2S
    V27_2S
    V28_2S
    V29_2S
    V30_2S
    V31_2S
)

const (
    V0_4S = SIMDRegister128v(Vec4S << 5) | iota
    V1_4S
    V2_4S
    V3_4S
    V4_4S
    V5_4S
    V6_4S
    V7_4S
    V8_4S
    V9_4S
    V10_4S
    V11_4S
    V12_4S
    V13_4S
    V14_4S
    V15_4S
    V16_4S
    V17_4S
    V18_4S
    V19_4S
    V20_4S
    V21_4S
    V22_4S
    V23_4S
    V24_4S
    V25_4S
    V26_4S
    V27_4S
    V28_4S
    V29_4S
    V30_4S
    V31_4S
)

const (
    V0_1D = SIMDRegister128v(Vec1D << 5) | iota
    V1_1D
    V2_1D
    V3_1D
    V4_1D
    V5_1D
    V6_1D
    V7_1D
    V8_1D
    V9_1D
    V10_1D
    V11_1D
    V12_1D
    V13_1D
    V14_1D
    V15_1D
    V16_1D
    V17_1D
    V18_1D
    V19_1D
    V20_1D
    V21_1D
    V22_1D
    V23_1D
    V24_1D
    V25_1D
    V26_1D
    V27_1D
    V28_1D
    V29_1D
    V30_1D
    V31_1D
)

const (
    V0_2D = SIMDRegister128v(Vec2D << 5) | iota
    V1_2D
    V2_2D
    V3_2D
    V4_2D
    V5_2D
    V6_2D
    V7_2D
    V8_2D
    V9_2D
    V10_2D
    V11_2D
    V12_2D
    V13_2D
    V14_2D
    V15_2D
    V16_2D
    V17_2D
    V18_2D
    V19_2D
    V20_2D
    V21_2D
    V22_2D
    V23_2D
    V24_2D
    V25_2D
    V26_2D
    V27_2D
    V28_2D
    V29_2D
    V30_2D
    V31_2D
)
