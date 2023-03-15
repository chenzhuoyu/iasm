package aarch64

import (
    `fmt`
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Vector represents a SIMD vector with a certain number of elements.
type Vector interface {
    asm.Register
    Size() uint8
    Format() VecFormat
    Vector()
}

type (
    _Vec1 uint16
    _Vec2 uint16
    _Vec3 uint16
    _Vec4 uint16
)

func mkvec(r VRegister) uint16 {
    return (uint16(r.Format()) << 8) | uint16(r.ID())
}

func checkvec(v ...VRegister) {
    for i, x := range v[1:] {
        if x.ID() != (v[i].ID() + 1) % 32 {
            panic("aarch64: vector elements must be consecutive registers")
        } else if x.Format() != v[i].Format() {
            panic("aarch64: vector elements must have identical arrangements")
        }
    }
}

func formatvec(v uint8, f VecFormat, n int) string {
    i := 0
    s := make([]string, 0, n)

    /* dump each element */
    for i < n {
        s = append(s, fmt.Sprintf("v%d.%s", (v + uint8(i)) % 32, f))
        i++
    }

    /* join them together */
    return fmt.Sprintf(
        "{ %s }",
        strings.Join(s, ", "),
    )
}

// Vec1 creates a vector with a single element.
func Vec1(v0 VRegister) Vector {
    return _Vec1(mkvec(v0))
}

// Vec2 creates a vector with two elements of the same arrangement.
func Vec2(v0, v1 VRegister) Vector {
    checkvec(v0, v1)
    return _Vec2(mkvec(v0))
}

// Vec3 creates a vector with 3 elements of the same arrangement.
func Vec3(v0, v1, v2 VRegister) Vector {
    checkvec(v0, v1, v2)
    return _Vec3(mkvec(v0))
}

// Vec4 creates a vector with 4 elements of the same arrangement.
func Vec4(v0, v1, v2, v3 VRegister) Vector {
    checkvec(v0, v1, v2, v3)
    return _Vec4(mkvec(v0))
}

func (_Vec1) Size() uint8 { return 1 }
func (_Vec2) Size() uint8 { return 2 }
func (_Vec3) Size() uint8 { return 3 }
func (_Vec4) Size() uint8 { return 4 }

func (_Vec1) Vector() {}
func (_Vec2) Vector() {}
func (_Vec3) Vector() {}
func (_Vec4) Vector() {}

func (_Vec1) Sealed(tag.Tag) {}
func (_Vec2) Sealed(tag.Tag) {}
func (_Vec3) Sealed(tag.Tag) {}
func (_Vec4) Sealed(tag.Tag) {}

func (self _Vec1) ID() uint8 { return uint8(self & 0x1f) }
func (self _Vec2) ID() uint8 { return uint8(self & 0x1f) }
func (self _Vec3) ID() uint8 { return uint8(self & 0x1f) }
func (self _Vec4) ID() uint8 { return uint8(self & 0x1f) }

func (self _Vec1) String() string { return formatvec(self.ID(), self.Format(), 1) }
func (self _Vec2) String() string { return formatvec(self.ID(), self.Format(), 2) }
func (self _Vec3) String() string { return formatvec(self.ID(), self.Format(), 3) }
func (self _Vec4) String() string { return formatvec(self.ID(), self.Format(), 4) }

func (self _Vec1) Format() VecFormat { return VecFormat(self >> 8) }
func (self _Vec2) Format() VecFormat { return VecFormat(self >> 8) }
func (self _Vec3) Format() VecFormat { return VecFormat(self >> 8) }
func (self _Vec4) Format() VecFormat { return VecFormat(self >> 8) }

// IndexedVector represents an indexed SIMD vector.
type IndexedVector interface {
    asm.Register
    Size() uint8
    Index() uint8
    IndexMode() VecIndexMode
    IndexedVector()
}

type (
    _IndexedVec1 uint16
    _IndexedVec2 uint16
    _IndexedVec3 uint16
    _IndexedVec4 uint16
)

func mkidx(v VidxRegister, i uint8) uint16 {
    return (uint16(i) << 8) | (uint16(v.IndexMode()) << 5) | uint16(v.ID())
}

func checkidx(v ...VidxRegister) {
    for i, x := range v[1:] {
        if x.ID() != (v[i].ID() + 1) % 32 {
            panic("aarch64: indexed vector elements must be consecutive registers")
        } else if x.IndexMode() != v[i].IndexMode() {
            panic("aarch64: indexed vector elements must have identical indexing modes")
        }
    }
}

func formatidx(v uint8, m VecIndexMode, n int, x uint8) string {
    i := 0
    s := make([]string, 0, n)

    /* dump each element */
    for i < n {
        s = append(s, fmt.Sprintf("v%d.%s", (v + uint8(i)) % 32, m))
        i++
    }

    /* join them together */
    return fmt.Sprintf(
        "{ %s }[%d]",
        strings.Join(s, ", "),
        x,
    )
}

// Index1 creates an indexed vector with a single element.
func Index1(v0 VidxRegister, i uint8) IndexedVector {
    return _IndexedVec1(mkidx(v0, i))
}

// Index2 creates an indexed vector with two elements of the same structure.
func Index2(v0, v1 VidxRegister, i uint8) IndexedVector {
    checkidx(v0, v1)
    return _IndexedVec2(mkidx(v0, i))
}

// Index3 creates an indexed vector with 3 elements of the same structure.
func Index3(v0, v1, v2 VidxRegister, i uint8) IndexedVector {
    checkidx(v0, v1, v2)
    return _IndexedVec3(mkidx(v0, i))
}

// Index4 creates an indexed vector with 4 elements of the same structure.
func Index4(v0, v1, v2, v3 VidxRegister, i uint8) IndexedVector {
    checkidx(v0, v1, v2, v3)
    return _IndexedVec4(mkidx(v0, i))
}

func (_IndexedVec1) Size() uint8 { return 1 }
func (_IndexedVec2) Size() uint8 { return 2 }
func (_IndexedVec3) Size() uint8 { return 3 }
func (_IndexedVec4) Size() uint8 { return 4 }

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

func (self _IndexedVec1) String() string { return formatidx(self.ID(), self.IndexMode(), 1, self.Index()) }
func (self _IndexedVec2) String() string { return formatidx(self.ID(), self.IndexMode(), 2, self.Index()) }
func (self _IndexedVec3) String() string { return formatidx(self.ID(), self.IndexMode(), 3, self.Index()) }
func (self _IndexedVec4) String() string { return formatidx(self.ID(), self.IndexMode(), 4, self.Index()) }

func (self _IndexedVec1) IndexMode() VecIndexMode { return VecIndexMode(uint8(self) >> 5) }
func (self _IndexedVec2) IndexMode() VecIndexMode { return VecIndexMode(uint8(self) >> 5) }
func (self _IndexedVec3) IndexMode() VecIndexMode { return VecIndexMode(uint8(self) >> 5) }
func (self _IndexedVec4) IndexMode() VecIndexMode { return VecIndexMode(uint8(self) >> 5) }
