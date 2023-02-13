package aarch64

import (
    `github.com/chenzhuoyu/iasm/internal/rt`
)

func asImm12    (v interface{}) uint32 { x, _ := asInt64(v)  ; return uint32(x & 0b111111111111) }
func asUimm4    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b1111) }
func asUimm6    (v interface{}) uint32 { x, _ := asUint64(v) ; return uint32(x & 0b111111) }
func asMaskImm  (v interface{}) uint32 { x, _ := asUint64(v) ; return _BitMask(x).mask() }

func asInt64(v interface{}) (int64, bool) {
    if isSpecial(v) {
        return 0, false
    } else if x := rt.AsEface(v); isInt(x.Kind()) {
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
    } else {
        return 0, false
    }
}
