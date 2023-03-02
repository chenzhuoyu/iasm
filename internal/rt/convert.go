package rt

import (
    `unsafe`
)

var (
    _u16tab [1 << 16]uint16
)

func init() {
    for i := 0; i < 1 << 16; i++ {
        _u16tab[i] = uint16(i)
    }
}

func WrapUint16(t *GoType, v uint16) interface{} {
    r := GoEface { Ty: t, Ptr: unsafe.Pointer(&_u16tab[v]) }
    return r.Pack()
}
