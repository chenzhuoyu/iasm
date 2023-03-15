package aarch64

import (
    `unsafe`
)

func bfv(v uint32, i int) uint32 {
    return v >> i
}

func mask(v uint32, n int) uint32 {
    return v & ((1 << n) - 1)
}

func maskp(v uint32, p int, n int) uint32 {
    return mask(bfv(v, p), n)
}

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