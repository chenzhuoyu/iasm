package aarch64

import (
    `fmt`
    `math`
    `testing`

    `github.com/stretchr/testify/require`
)

func TestBitMask_Encode(t *testing.T) {
    testtab := []struct {
        imm           uint64
        n, immr, imms uint32
        ok            bool
    } {
        { imm: 0x5555555555555555, n: 0, immr: 0b000000, imms: 0b111100, ok: true },
        { imm: 0xaaaaaaaaaaaaaaaa, n: 0, immr: 0b000001, imms: 0b111100, ok: true },
        { imm: 0x1111111111111111, n: 0, immr: 0b000000, imms: 0b111000, ok: true },
        { imm: 0x6666666666666666, n: 0, immr: 0b000011, imms: 0b111001, ok: true },
        { imm: 0xeeeeeeeeeeeeeeee, n: 0, immr: 0b000011, imms: 0b111010, ok: true },
        { imm: 0x0101010101010101, n: 0, immr: 0b000000, imms: 0b110000, ok: true },
        { imm: 0x1818181818181818, n: 0, immr: 0b000101, imms: 0b110001, ok: true },
        { imm: 0xfefefefefefefefe, n: 0, immr: 0b000111, imms: 0b110110, ok: true },
        { imm: 0x0001000100010001, n: 0, immr: 0b000000, imms: 0b100000, ok: true },
        { imm: 0xff8fff8fff8fff8f, n: 0, immr: 0b001001, imms: 0b101100, ok: true },
        { imm: 0xfffefffefffefffe, n: 0, immr: 0b001111, imms: 0b101110, ok: true },
        { imm: 0x0000000100000001, n: 0, immr: 0b000000, imms: 0b000000, ok: true },
        { imm: 0x3fffff003fffff00, n: 0, immr: 0b011000, imms: 0b010101, ok: true },
        { imm: 0xfffffffefffffffe, n: 0, immr: 0b011111, imms: 0b011110, ok: true },
        { imm: 0x0000000000000001, n: 1, immr: 0b000000, imms: 0b000000, ok: true },
        { imm: 0x0000001fffff0000, n: 1, immr: 0b110000, imms: 0b010100, ok: true },
        { imm: 0xfffffffffffffffe, n: 1, immr: 0b111111, imms: 0b111110, ok: true },
        { imm:  5, ok: false },
        { imm:  9, ok: false },
        { imm: 10, ok: false },
        { imm: 11, ok: false },
        { imm: 13, ok: false },
        { imm: 17, ok: false },
        { imm: 18, ok: false },
        { imm: 19, ok: false },
        { imm: math.MaxUint64, ok: false },
    }
    for _, v := range testtab {
        t.Run(fmt.Sprintf("%016x", v.imm), func(t *testing.T) {
            n, immr, imms, ok := _BitMask(v.imm).encode()
            require.Equalf(t, v.ok, ok, "ok")
            require.Equalf(t, v.n, n, "n")
            require.Equalf(t, v.immr, immr, "immr")
            require.Equalf(t, v.imms, imms, "imms")
        })
    }
}
