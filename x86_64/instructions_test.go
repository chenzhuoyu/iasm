package x86_64

import (
    `testing`

    `github.com/davecgh/go-spew/spew`
)

func TestEncodeInstr(t *testing.T) {
    a := CreateArch()
    p := a.CreateProgram()
    v := p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).Encode()
    spew.Dump(v)
}

func BenchmarkEncodeInstr(b *testing.B) {
    a := CreateArch()
    m := make([]byte, 0, 16)
    p := a.CreateProgram()
    p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).EncodeInto(&m)
    p.Free()
    b.SetBytes(int64(len(m)))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        m = m[:0]
        p = a.CreateProgram()
        p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).EncodeInto(&m)
        p.Free()
    }
}
