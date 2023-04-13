package aarch64

import (
    `math`
)

func encodeFpImm8(v interface{}) (uint8, bool) {
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
    m := x & 0x0f_ffff_ffff_ffff

    /* check mantissa range */
    if m & 0xffff_ffff_ffff != 0 {
        return 0, false
    } else {
        return uint8((s << 7) | (e << 4) | (m >> 48)), true
    }
}
