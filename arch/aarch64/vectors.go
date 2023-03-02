package aarch64

import (
    `fmt`

    `github.com/chenzhuoyu/iasm/internal/rt`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// SIMDVector1 represents a SIMD vector with a single register.
type SIMDVector1 uint8

// SIMDVector2 represents a SIMD vector with two registers of the same arrangement.
type SIMDVector2 uint8

// SIMDVector3 represents a SIMD vector with three registers of the same arrangement.
type SIMDVector3 uint8

// SIMDVector4 represents a SIMD vector with four registers of the same arrangement.
type SIMDVector4 uint8

func nextreg(v, i SIMDRegister128v) SIMDRegister128v {
    return (v + i) % 32
}

func checkvec(v ...SIMDRegister128v) {
    for i, x := range v[1:] {
        if x != nextreg(v[i], 1) {
            panic("aarch64: vector elements must be consecutive registers")
        } else if x.Arrangement() != v[i].Arrangement() {
            panic("aarch64: vector elements must have identical arrangements")
        }
    }
}

func formatvec(v SIMDRegister128v, n int) string {
    switch n {
        case 1  : return fmt.Sprintf("{ %s }", v)
        case 2  : return fmt.Sprintf("{ %s, %s }", v, nextreg(v, 1))
        case 3  : return fmt.Sprintf("{ %s, %s, %s }", v, nextreg(v, 1), nextreg(v, 2))
        case 4  : return fmt.Sprintf("{ %s, %s, %s, %s }", v, nextreg(v, 1), nextreg(v, 2), nextreg(v, 3))
        default : panic("unreachable")
    }
}

// Vec1 creates a vector with a single element.
func Vec1(v0 SIMDRegister128v) SIMDVector1 {
    return SIMDVector1(v0)
}

// Vec2 creates a vector with two elements of the same arrangement.
func Vec2(v0, v1 SIMDRegister128v) SIMDVector2 {
    checkvec(v0, v1)
    return SIMDVector2(v0)
}

// Vec3 creates a vector with 3 elements of the same arrangement.
func Vec3(v0, v1, v2 SIMDRegister128v) SIMDVector3 {
    checkvec(v0, v1, v2)
    return SIMDVector3(v0)
}

// Vec4 creates a vector with 4 elements of the same arrangement.
func Vec4(v0, v1, v2, v3 SIMDRegister128v) SIMDVector4 {
    checkvec(v0, v1, v2, v3)
    return SIMDVector4(v0)
}

func (self SIMDVector1) ID() uint8 { return uint8(self & 0x1f) }
func (self SIMDVector2) ID() uint8 { return uint8(self & 0x1f) }
func (self SIMDVector3) ID() uint8 { return uint8(self & 0x1f) }
func (self SIMDVector4) ID() uint8 { return uint8(self & 0x1f) }

func (self SIMDVector1) String() string { return formatvec(SIMDRegister128v(self), 1) }
func (self SIMDVector2) String() string { return formatvec(SIMDRegister128v(self), 2) }
func (self SIMDVector3) String() string { return formatvec(SIMDRegister128v(self), 3) }
func (self SIMDVector4) String() string { return formatvec(SIMDRegister128v(self), 4) }

func (self SIMDVector1) Sealed(tag.Tag) {}
func (self SIMDVector2) Sealed(tag.Tag) {}
func (self SIMDVector3) Sealed(tag.Tag) {}
func (self SIMDVector4) Sealed(tag.Tag) {}

// SIMDIndexedVector represents an indexed SIMD vector.
type SIMDIndexedVector interface {
    tag.Sealed
    ID() uint8
    Index() uint8
    Structure() SIMDVectorStructure
    IndexedVector()
}

type (
    _IndexedVec1 uint16
    _IndexedVec2 uint16
    _IndexedVec3 uint16
    _IndexedVec4 uint16
)

var (
    _typeIndexedVec1 = rt.TypeOf(_IndexedVec1(0))
    _typeIndexedVec2 = rt.TypeOf(_IndexedVec2(0))
    _typeIndexedVec3 = rt.TypeOf(_IndexedVec3(0))
    _typeIndexedVec4 = rt.TypeOf(_IndexedVec4(0))
)

func nextidx(v, i SIMDRegister128r) SIMDRegister128r {
    return (v + i) % 32
}

func checkidx(v ...SIMDRegister128r) {
    for i, x := range v[1:] {
        if x != nextidx(v[i], 1) {
            panic("aarch64: indexed vector elements must be consecutive registers")
        } else if x.Structure() != v[i].Structure() {
            panic("aarch64: indexed vector elements must have identical structures")
        }
    }
}

func formatidx(v SIMDRegister128r, n int, i uint8) string {
    switch n {
        case 1  : return fmt.Sprintf("{ %s }[%d]", v, i)
        case 2  : return fmt.Sprintf("{ %s, %s }[%d]", v, nextidx(v, 1), i)
        case 3  : return fmt.Sprintf("{ %s, %s, %s }[%d]", v, nextidx(v, 1), nextidx(v, 2), i)
        case 4  : return fmt.Sprintf("{ %s, %s, %s, %s }[%d]", v, nextidx(v, 1), nextidx(v, 2), nextidx(v, 3), i)
        default : panic("unreachable")
    }
}

// Index1 creates an indexed vector with a single element.
func Index1(v0 SIMDRegister128r, i uint8) SIMDIndexedVector {
    return v0.withIndex(_typeIndexedVec1, i).(SIMDIndexedVector)
}

// Index2 creates an indexed vector with two elements of the same structure.
func Index2(v0, v1 SIMDRegister128r, i uint8) SIMDIndexedVector {
    checkidx(v0, v1)
    return v0.withIndex(_typeIndexedVec2, i).(SIMDIndexedVector)
}

// Index3 creates an indexed vector with 3 elements of the same structure.
func Index3(v0, v1, v2 SIMDRegister128r, i uint8) SIMDIndexedVector {
    checkidx(v0, v1, v2)
    return v0.withIndex(_typeIndexedVec3, i).(SIMDIndexedVector)
}

// Index4 creates an indexed vector with 4 elements of the same structure.
func Index4(v0, v1, v2, v3 SIMDRegister128r, i uint8) SIMDIndexedVector {
    checkidx(v0, v1, v2, v3)
    return v0.withIndex(_typeIndexedVec4, i).(SIMDIndexedVector)
}

func (_IndexedVec1) Sealed(tag.Tag) {}
func (_IndexedVec2) Sealed(tag.Tag) {}
func (_IndexedVec3) Sealed(tag.Tag) {}
func (_IndexedVec4) Sealed(tag.Tag) {}

func (_IndexedVec1) IndexedVector() {}
func (_IndexedVec2) IndexedVector() {}
func (_IndexedVec3) IndexedVector() {}
func (_IndexedVec4) IndexedVector() {}

func (self _IndexedVec1) ID() uint8 { return uint8(self & 0x1f) }
func (self _IndexedVec2) ID() uint8 { return uint8(self & 0x1f) }
func (self _IndexedVec3) ID() uint8 { return uint8(self & 0x1f) }
func (self _IndexedVec4) ID() uint8 { return uint8(self & 0x1f) }

func (self _IndexedVec1) Index() uint8 { return uint8(self >> 8) }
func (self _IndexedVec2) Index() uint8 { return uint8(self >> 8) }
func (self _IndexedVec3) Index() uint8 { return uint8(self >> 8) }
func (self _IndexedVec4) Index() uint8 { return uint8(self >> 8) }

func (self _IndexedVec1) String() string { return formatidx(SIMDRegister128r(self), 1, self.Index()) }
func (self _IndexedVec2) String() string { return formatidx(SIMDRegister128r(self), 2, self.Index()) }
func (self _IndexedVec3) String() string { return formatidx(SIMDRegister128r(self), 3, self.Index()) }
func (self _IndexedVec4) String() string { return formatidx(SIMDRegister128r(self), 4, self.Index()) }

func (self _IndexedVec1) Structure() SIMDVectorStructure { return SIMDVectorStructure(uint8(self) >> 5) }
func (self _IndexedVec2) Structure() SIMDVectorStructure { return SIMDVectorStructure(uint8(self) >> 5) }
func (self _IndexedVec3) Structure() SIMDVectorStructure { return SIMDVectorStructure(uint8(self) >> 5) }
func (self _IndexedVec4) Structure() SIMDVectorStructure { return SIMDVectorStructure(uint8(self) >> 5) }
