package aarch64

import (
    `math`

    `github.com/chenzhuoyu/iasm/internal/rt`
)

func asInt32    (v interface{}) int32  { x, _ := asInt64(v)  ; return int32(x) }
func asImm12    (v interface{}) uint32 { x, _ := asInt64(v)  ; return uint32(x & 0b111111111111) }
func asUimm4    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b1111) }
func asUimm5    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b11111) }
func asUimm6    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b111111) }
func asUimm8    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0xff) }
func asUimm16   (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0xffff) }
func asMaskOp   (v interface{}) uint32 { x, _ := asUint64(v) ; return _BitMask(x).mask() }
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
