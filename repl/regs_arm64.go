package repl

import (
    `fmt`
)

type _RegFileARM64 struct {
    x    [32]uint64
    v    [32][16]byte
    pc   uint64
    nzcv uint64
}

//go:nosplit
//go:noescape
//goland:noinspection GoUnusedParameter
func dumpregs(regs *_RegFileARM64)

func (self *_RegFileARM64) Dump(indent int) []_Reg {
    dumpregs(self)
    return self.format(indent)
}

func (self *_RegFileARM64) Compare(other _RegFile, indent int) (v []_Reg) {
    var ok bool
    var rf *_RegFileARM64

    /* must be an AMD64 register file */
    if rf, ok = other.(*_RegFileARM64); !ok {
        panic("can only compare with ARM64 register file")
    }

    /* generic registers (except fp, lr & sp) */
    for i := 0; i < 29; i++ {
        if self.x[i] != rf.x[i] {
            v = append(v, _Reg {
                reg: fmt.Sprintf("x%d", i),
                val: fmt.Sprintf("%#016x -> %#016x", self.x[i], rf.x[i]),
                vec: false,
            })
        }
    }

    /* fp, lr & sp, pc is always going to change, so ignore here */
    if self.x[29] != rf.x[29] { v = append(v, _Reg { reg: "fp", val: fmt.Sprintf("%#016x -> %#016x", self.x[29], rf.x[29]), vec: false }) }
    if self.x[30] != rf.x[30] { v = append(v, _Reg { reg: "lr", val: fmt.Sprintf("%#016x -> %#016x", self.x[30], rf.x[30]), vec: false }) }
    if self.x[31] != rf.x[31] { v = append(v, _Reg { reg: "sp", val: fmt.Sprintf("%#016x -> %#016x", self.x[31], rf.x[31]), vec: false }) }

    /* vector registers */
    for i, x := range self.v {
        if x != rf.v[i] {
            v = append(v, _Reg {
                reg: fmt.Sprintf("v%d", i),
                val: vecdiff(x[:], rf.v[i][:], indent),
                vec: true,
            })
        }
    }

    /* check flags register */
    if self.nzcv == rf.nzcv {
        return v
    }

    /* flags changed */
    return append(v, _Reg {
        reg: "nzcv",
        val: fmt.Sprintf("%s -> %s", nzcv(self.nzcv), nzcv(rf.nzcv)),
        vec: false,
    })
}

func (self *_RegFileARM64) format(indent int) (v []_Reg) {
    self.formatStd(&v)
    self.formatExt(&v)
    self.formatVec(&v, indent)
    self.formatNZCV(&v)
    return
}

func (self *_RegFileARM64) formatStd(v *[]_Reg) {
    for i := 0; i < 29; i++ {
        *v = append(*v, _Reg {
            reg: fmt.Sprintf("x%d", i),
            val: fmt.Sprintf("%#016x %s", self.x[i], symbol(self.x[i])),
            vec: false,
        })
    }
}

func (self *_RegFileARM64) formatExt(v *[]_Reg) {
    *v = append(
        *v,
        _Reg { reg: "fp", val: fmt.Sprintf("%#016x %s", self.x[29], symbol(self.x[29])), vec: false },
        _Reg { reg: "lr", val: fmt.Sprintf("%#016x %s", self.x[30], symbol(self.x[30])), vec: false },
        _Reg { reg: "sp", val: fmt.Sprintf("%#016x %s", self.x[31], symbol(self.x[31])), vec: false },
        _Reg { reg: "pc", val: fmt.Sprintf("%#016x %s", self.pc, symbol(self.pc)), vec: false },
    )
}

func (self *_RegFileARM64) formatVec(v *[]_Reg, indent int) {
    for i, x := range self.v {
        *v = append(*v, _Reg {
            reg: fmt.Sprintf("v%d", i),
            val: vector(x[:], indent),
            vec: true,
        })
    }
}

func (self *_RegFileARM64) formatNZCV(v *[]_Reg) {
    *v = append(*v, _Reg {
        reg: "nzcv",
        val: nzcv(self.nzcv),
        vec: false,
    })
}

func init() {
    _regs = new(_RegFileARM64)
}
