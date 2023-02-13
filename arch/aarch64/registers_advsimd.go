package aarch64

import (
    `fmt`
    
    `github.com/chenzhuoyu/iasm/internal/tag`
)

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
        case Vec8B  : return "8b"
        case Vec16B : return "16b"
        case Vec4H  : return "4h"
        case Vec8H  : return "8h"
        case Vec2S  : return "2s"
        case Vec4S  : return "4s"
        case Vec1D  : return "1d"
        case Vec2D  : return "2d"
        default     : panic("aarch64: invalid SIMD vector arrangement")
    }
}

// SIMDVector1 represents an unarranged SIMD vector with a single register.
type SIMDVector1 [1]SIMDRegister128v

// SIMDVector2 represents an unarranged SIMD vector with two registers that share the same arrangement.
type SIMDVector2 [2]SIMDRegister128v

// SIMDVector3 represents an unarranged SIMD vector with three registers that share the same arrangement.
type SIMDVector3 [3]SIMDRegister128v

// SIMDVector4 represents an unarranged SIMD vector with four registers that share the same arrangement.
type SIMDVector4 [4]SIMDRegister128v

func Vec1(v0             SIMDRegister128v) SIMDVector1 { return SIMDVector1 { v0             } }
func Vec2(v0, v1         SIMDRegister128v) SIMDVector2 { return SIMDVector2 { v0, v1         } }
func Vec3(v0, v1, v2     SIMDRegister128v) SIMDVector3 { return SIMDVector3 { v0, v1, v2     } }
func Vec4(v0, v1, v2, v3 SIMDRegister128v) SIMDVector4 { return SIMDVector4 { v0, v1, v2, v3 } }

func (self SIMDVector1) As(v SIMDVectorArrangement) SIMDVector1r { return SIMDVector1r { self, v } }
func (self SIMDVector2) As(v SIMDVectorArrangement) SIMDVector2r { return SIMDVector2r { self, v } }
func (self SIMDVector3) As(v SIMDVectorArrangement) SIMDVector3r { return SIMDVector3r { self, v } }
func (self SIMDVector4) As(v SIMDVectorArrangement) SIMDVector4r { return SIMDVector4r { self, v } }

func (self SIMDVector1) String() string {
    return fmt.Sprintf("{ %s }", self[0])
}

func (self SIMDVector2) String() string {
    return fmt.Sprintf("{ %s, %s }", self[0], self[1])
}

func (self SIMDVector3) String() string {
    return fmt.Sprintf("{ %s, %s, %s }", self[0], self[1], self[2])
}

func (self SIMDVector4) String() string {
    return fmt.Sprintf("{ %s, %s, %s, %s }", self[0], self[1], self[2], self[3])
}

type (
    SIMDVector1r struct { V SIMDVector1; A SIMDVectorArrangement }
    SIMDVector2r struct { V SIMDVector2; A SIMDVectorArrangement }
    SIMDVector3r struct { V SIMDVector3; A SIMDVectorArrangement }
    SIMDVector4r struct { V SIMDVector4; A SIMDVectorArrangement }
)

func (self SIMDVector1r) String() string {
    return fmt.Sprintf("{ %s }.%s", self.V[0], self.A)
}

func (self SIMDVector2r) String() string {
    return fmt.Sprintf("{ %s, %s }.%s", self.V[0], self.V[1], self.A)
}

func (self SIMDVector3r) String() string {
    return fmt.Sprintf("{ %s, %s, %s }.%s", self.V[0], self.V[1], self.V[2], self.A)
}

func (self SIMDVector4r) String() string {
    return fmt.Sprintf("{ %s, %s, %s, %s }.%s", self.V[0], self.V[1], self.V[2], self.V[3], self.A)
}

type (
    SIMDRegister8    uint8
    SIMDRegister16   uint8
    SIMDRegister32   uint8
    SIMDRegister64   uint8
    SIMDRegister128  uint8
    SIMDRegister128v uint8
    SIMDRegister128r uint8
)

func (SIMDRegister8)    Sealed(tag.Tag) {}
func (SIMDRegister16)   Sealed(tag.Tag) {}
func (SIMDRegister32)   Sealed(tag.Tag) {}
func (SIMDRegister64)   Sealed(tag.Tag) {}
func (SIMDRegister128)  Sealed(tag.Tag) {}
func (SIMDRegister128r) Sealed(tag.Tag) {}

func (self SIMDRegister8)    ID() uint8 { return uint8(self) }
func (self SIMDRegister16)   ID() uint8 { return uint8(self) }
func (self SIMDRegister32)   ID() uint8 { return uint8(self) }
func (self SIMDRegister64)   ID() uint8 { return uint8(self) }
func (self SIMDRegister128)  ID() uint8 { return uint8(self) }
func (self SIMDRegister128r) ID() uint8 { return uint8(self) }

func (self SIMDRegister8)    String() string { return fmt.Sprintf("b%d", self.ID()) }
func (self SIMDRegister16)   String() string { return fmt.Sprintf("h%d", self.ID()) }
func (self SIMDRegister32)   String() string { return fmt.Sprintf("s%d", self.ID()) }
func (self SIMDRegister64)   String() string { return fmt.Sprintf("d%d", self.ID()) }
func (self SIMDRegister128)  String() string { return fmt.Sprintf("q%d", self.ID()) }
func (self SIMDRegister128v) String() string { return fmt.Sprintf("v%d", self) }
func (self SIMDRegister128r) String() string { return fmt.Sprintf("v%d.%s", self.ID(), self.Arrangement()) }

func (self SIMDRegister128v) As(v SIMDVectorArrangement) SIMDRegister128r {
    if self &^ 0b11111 != 0 {
        panic("aarch64: invalid unarranged vector register")
    } else {
        return SIMDRegister128r(uint8(v) << 5 | uint8(self))
    }
}

func (self SIMDRegister128r) Arrangement() SIMDVectorArrangement {
    return SIMDVectorArrangement(self >> 5)
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
    V0 SIMDRegister128v = iota
    V1
    V2
    V3
    V4
    V5
    V6
    V7
    V8
    V9
    V10
    V11
    V12
    V13
    V14
    V15
    V16
    V17
    V18
    V19
    V20
    V21
    V22
    V23
    V24
    V25
    V26
    V27
    V28
    V29
    V30
    V31
)
