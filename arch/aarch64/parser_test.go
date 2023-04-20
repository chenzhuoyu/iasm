package aarch64

import (
    `strings`
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/stretchr/testify/require`
)

func TestParser_Parse(t *testing.T) {
    p := asm.GetArch("aarch64").CreateParser()
    v, e := p.Feed("LDR " + strings.Join([]string {
        `#0`,
        `#123`,
        `#0.0`,
        `#0.5`,
        `#1.0`,
        `#2.0`,
        `x0`,
        `v0.16B`,
        `v5.2H`,
        `v7.2H[3]`,
        `v9.B[7]`,
        `v3.b[5]`,
        `RCTX`,
        `VMALLE1NXS`,
        `DBGWVR5_EL1`,
        `UAO`,
        `PAN`,
        `CurrentEL`,
        `DAIFSet`,
        `SVCRSM`,
        `{ v31.2h, v0.2h, v1.2h, v2.2h }`,
        `{ v0.b, v1.b, v2.b, v3.b }[15]`,
    }, ", "))
    require.NoError(t, e)
    println(v.String())
}
