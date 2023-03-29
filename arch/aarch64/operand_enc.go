package aarch64

import (
    `math`

    `github.com/chenzhuoyu/iasm/internal/rt`
)

func asInt32    (v interface{}) int32  { x, _ := asInt64(v)  ; return int32(x) }
func asImm12    (v interface{}) uint32 { x, _ := asInt64(v)  ; return uint32(x & 0b111111111111) }
func asUimm3    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b111) }
func asUimm4    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b1111) }
func asUimm5    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b11111) }
func asUimm6    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b111111) }
func asUimm7    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b1111111) }
func asUimm8    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0xff) }
func asUimm16   (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0xffff) }
func asMaskOp   (v interface{}) uint32 { x, _ := asUint64(v) ; return _BitMask(x).mask() }
func asFpImm8   (v interface{}) uint32 { x, _ := asFloat8(v) ; return uint32(x) }
func asFpScale  (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(64 - (x & 0b111111)) }

func asLit(v interface{}) int64 {
    if isSpecial(v) {
        return 0
    } else if x := rt.AsEface(v); !isInt(x.Kind()) && !isUint(x.Kind()) {
        return 0
    } else {
        return x.ToInt64()
    }
}

func asInt64(v interface{}) (int64, bool) {
    if isSpecial(v) {
        return 0, false
    } else if x := rt.AsEface(v); isInt(x.Kind()) {
        return x.ToInt64(), true
    } else if isUint(x.Kind()) && x.ToUint64() <= math.MaxInt64 {
        return x.ToInt64(), true
    } else {
        return 0, false
    }
}

func asUint64(v interface{}) (uint64, bool) {
    if isSpecial(v) {
        return 0, false
    } else if x := rt.AsEface(v); isUint(x.Kind()) {
        return x.ToUint64(), true
    } else if isInt(x.Kind()) && x.ToInt64() >= 0 {
        return x.ToUint64(), true
    } else {
        return 0, false
    }
}

func asFloat8(v interface{}) (uint8, bool) {
    var ok bool
    var fv float64

    /* convert to float64 */
    if fv, ok = asFloat64(v); !ok {
        return 0, false
    }

    /* extract the exponent */
    x := int64(math.Float64bits(fv))
    e := ((x >> 52) & 0x7ff) - 1023

    /* check exponent range */
    if e < -3 || e > 4 {
        return 0, false
    }

    /* extract sign and mantissa */
    s := (x >> 63) & 1
    m := x & 0xf_ffff_ffff_ffff

    /* check mantissa range */
    if m & 0xffff_ffff_ffff != 0 {
        return 0, false
    } else {
        return uint8((s << 7) | (e << 4) | (m >> 48)), true
    }
}

func asFloat64(v interface{}) (float64, bool) {
    if isSpecial(v) {
        return 0, false
    } else if x := rt.AsEface(v); isInt(x.Kind()) {
        return float64(x.ToInt64()), true
    } else if isUint(x.Kind()) {
        return float64(x.ToUint64()), true
    } else if isFloat(x.Kind()) {
        return x.ToFloat64(), true
    } else {
        return 0, false
    }
}

func asMOVxImm(v interface{}, bits uint64, inverted bool) uint32 {
    x, _ := encodeMoveImm(v, bits, inverted)
    return x
}

func asExtIndex(r, v interface{}) uint32 {
    switch vfmt(r) {
        case Vec8B  : return asUimm4(v) & 0b00111
        case Vec16B : return asUimm4(v) | 0b10000
        default     : panic("aarch64: invalid register for EXT instruction")
    }
}

func asLSLShift(v interface{}, n uint64) uint32 {
    sh, _ := asUint64(v)
    return uint32((-sh % n) << 6 | (n - sh - 1))
}
