package aarch64

import (
    `fmt`
    `testing`

    `github.com/stretchr/testify/require`
)

func TestMoveImm_Encode(t *testing.T) {
    testtab := []struct {
        imm      uint64
        bits     uint64
        inverted bool
        ret      uint32
        ok       bool
    } {
        { imm:                  0, bits: 32, inverted: true , ok: false },
        { imm:         0xffffffff, bits: 32, inverted: true , ret: 0b00_0000000000000000, ok: true },
        { imm:         0xfffffffe, bits: 32, inverted: true , ret: 0b00_0000000000000001, ok: true },
        { imm:         0xffffffee, bits: 32, inverted: true , ret: 0b00_0000000000010001, ok: true },
        { imm:         0xfffffeee, bits: 32, inverted: true , ret: 0b00_0000000100010001, ok: true },
        { imm:         0xffffeeee, bits: 32, inverted: true , ret: 0b00_0001000100010001, ok: true },
        { imm:         0xfffeeeee, bits: 32, inverted: true , ok: false },
        { imm:         0xfffeffff, bits: 32, inverted: true , ret: 0b01_0000000000000001, ok: true },
        { imm:         0xffeeffff, bits: 32, inverted: true , ret: 0b01_0000000000010001, ok: true },
        { imm:         0xfeeeffff, bits: 32, inverted: true , ret: 0b01_0000000100010001, ok: true },
        { imm:         0xeeeeffff, bits: 32, inverted: true , ret: 0b01_0001000100010001, ok: true },
        { imm:                  0, bits: 64, inverted: true , ok: false },
        { imm: 0xffffffffffffffff, bits: 64, inverted: true , ret: 0b00_0000000000000000, ok: true },
        { imm: 0xfffffffffffffffe, bits: 64, inverted: true , ret: 0b00_0000000000000001, ok: true },
        { imm: 0xffffffffffffffee, bits: 64, inverted: true , ret: 0b00_0000000000010001, ok: true },
        { imm: 0xffffffffffeeffff, bits: 64, inverted: true , ret: 0b01_0000000000010001, ok: true },
        { imm: 0xffffffeeffffffff, bits: 64, inverted: true , ret: 0b10_0000000000010001, ok: true },
        { imm: 0xffeeffffffffffff, bits: 64, inverted: true , ret: 0b11_0000000000010001, ok: true },
        { imm:                  0, bits: 32, inverted: false, ret: 0b00_0000000000000000, ok: true },
        { imm:                  1, bits: 32, inverted: false, ret: 0b00_0000000000000001, ok: true },
        { imm:             0xffff, bits: 32, inverted: false, ret: 0b00_1111111111111111, ok: true },
        { imm:         0xffff0000, bits: 32, inverted: false, ret: 0b01_1111111111111111, ok: true },
        { imm:     0xffff00000000, bits: 32, inverted: false, ok: false },
        { imm: 0xffff000000000000, bits: 32, inverted: false, ok: false },
        { imm:                  0, bits: 64, inverted: false, ret: 0b00_0000000000000000, ok: true },
        { imm:                  1, bits: 64, inverted: false, ret: 0b00_0000000000000001, ok: true },
        { imm:             0xffff, bits: 64, inverted: false, ret: 0b00_1111111111111111, ok: true },
        { imm:         0xffff0000, bits: 64, inverted: false, ret: 0b01_1111111111111111, ok: true },
        { imm:     0xffff00000000, bits: 64, inverted: false, ret: 0b10_1111111111111111, ok: true },
        { imm: 0xffff000000000000, bits: 64, inverted: false, ret: 0b11_1111111111111111, ok: true },
        { imm: 0xffffffff00000000, bits: 64, inverted: false, ok: false },
    }
    for _, v := range testtab {
        t.Run(fmt.Sprintf("%016x", v.imm), func(t *testing.T) {
            ret, ok := encodeMoveImm(v.imm, v.bits, v.inverted)
            require.Equalf(t, v.ok, ok, "ok")
            require.Equalf(t, v.ret, ret, "ret")
        })
    }
}
