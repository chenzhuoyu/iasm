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
        `v7.2H[4]`,
        `v9.B[3]`,
        `v3.b[5]`,
        `RCTX`,
        `VMALLE1NXS`,
        `DBGWVR5_EL1`,
        `UAO`,
        `PAN`,
        `CurrentEL`,
        `DAIFSet`,
        `SVCRSM`,
    }, ", "))
    require.NoError(t, e)
    println(v.String())
}
