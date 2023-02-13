package aarch64

import (
    `math`
    `math/bits`
)

type (
    _BitMask uint64
)

func (self _BitMask) is32() bool {
    n, _, _, ok := self.encode()
    return ok && n == 0
}

func (self _BitMask) is64() bool {
    _, _, _, ok := self.encode()
    return ok
}

func (self _BitMask) mask() uint32 {
    if n, immr, imms, ok := self.encode(); !ok {
        panic("aarch64: invalid bit mask")
    } else {
        return n << 12 | immr << 6 | imms
    }
}

func (self _BitMask) encode() (n, immr, imms uint32, ok bool) {
    ok = false
    n, immr, imms = 0, 0, 0

    /* all 0s and all 1s cannot be encoded as bit mask */
    if self == 0 || self == math.MaxUint64 {
        return
    }

    /* calculate the rotation value and the normalized value */
    vals := uint64(self)
    rots := bits.TrailingZeros64(vals & (vals + 1))
    norm := bits.RotateLeft64(vals, -(rots & 0x3f))

    /* calculate pattern size */
    ld0s := bits.LeadingZeros64(norm)
    tr1s := bits.TrailingZeros64(^norm)
    size := ld0s + tr1s

    /* check if the value can be encoded */
    if ok = bits.RotateLeft64(vals, -(size & 0x3f)) == vals; !ok {
        return
    }

    /* encode the result */
    n    = uint32((size >> 6) & 1)
    immr = uint32(-rots & (size - 1)) & 0x3f
    imms = uint32(-(size << 1) | (tr1s - 1)) & 0x3f
    return
}
