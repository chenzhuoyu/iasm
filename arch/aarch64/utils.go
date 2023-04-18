package aarch64

import (
    `unsafe`

    `github.com/chenzhuoyu/iasm/asm`
)

func this(p *asm.Instruction) *Instruction {
    return (*Instruction)(unsafe.Pointer(p))
}

func mask(v uint32, n int) uint32 {
    return v & ((1 << n) - 1)
}

func ubfx(v uint32, p int, n int) uint32 {
    return mask(v >> p, n)
}

func u32at(p unsafe.Pointer, i int) uint32 {
    return *(*uint32)(unsafe.Pointer(uintptr(p) + uintptr(i) * 4))
}

func isident(ch rune) bool {
    return (ch >= '0' && ch <= '9') ||
           (ch >= 'A' && ch <= 'Z') ||
           (ch >= 'a' && ch <= 'z') ||
           (ch == '_')
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
