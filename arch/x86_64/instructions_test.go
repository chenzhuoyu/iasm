package x86_64

import (
    `testing`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/davecgh/go-spew/spew`
)

func TestInstr_Encode(t *testing.T) {
    m := []byte(nil)
    a := asm.GetArch("x86_64")
    p := a.CreateProgram().(*Program)
    p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).encode(&m)
    spew.Dump(m)
}

func TestInstr_EncodeSegment(t *testing.T) {
    m := []byte(nil)
    a := asm.GetArch("x86_64")
    p := a.CreateProgram().(*Program)
    p.MOVQ(Abs(0x30), RCX).GS().encode(&m)
    spew.Dump(m)
}

func BenchmarkInstr_Encode(b *testing.B) {
    m := make([]byte, 0, 16)
    a := asm.GetArch("x86_64")
    p := a.CreateProgram().(*Program)
    p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).encode(&m)
    p.Free()
    b.SetBytes(int64(len(m)))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        m = m[:0]
        p = a.CreateProgram().(*Program)
        p.VPERMIL2PD(7, Sib(R8, R9, 1, 12345), YMM1, YMM2, YMM3).encode(&m)
        p.Free()
    }
}
