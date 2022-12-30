package x86_64

import (
    `testing`

    `github.com/davecgh/go-spew/spew`
)

func TestInstr_Encode(t *testing.T) {
    m := []byte(nil)
    a := CreateArch()
    p := a.CreateProgram()
    p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).encode(&m)
    spew.Dump(m)
}

func TestInstr_EncodeSegment(t *testing.T) {
    m := []byte(nil)
    a := CreateArch()
    p := a.CreateProgram()
    p.MOVQ(Abs(0x30), RCX).GS().encode(&m)
    spew.Dump(m)
}

func BenchmarkInstr_Encode(b *testing.B) {
    a := CreateArch()
    m := make([]byte, 0, 16)
    p := a.CreateProgram()
    p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).encode(&m)
    p.Free()
    b.SetBytes(int64(len(m)))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        m = m[:0]
        p = a.CreateProgram()
        p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).encode(&m)
        p.Free()
    }
}
