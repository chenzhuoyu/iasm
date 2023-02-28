package aarch64

import (
    `unsafe`
)

func u32at(p unsafe.Pointer, i int) uint32 {
    return *(*uint32)(unsafe.Pointer(uintptr(p) + uintptr(i) * 4))
}

func matchany(v uint32, m *uint32, p *uint32, n int) (ret bool) {
    a := unsafe.Pointer(m)
    b := unsafe.Pointer(p)

    /* check if anything matches */
    for i := 0; i < n; i++ {
        if x := u32at(a, i); x != 0 && v & x == u32at(b, i) {
            return true
        }
    }

    /* matched nothing */
    ret = false
    return
}
