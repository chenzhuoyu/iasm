package repl

type _Reg struct {
    reg string
    val string
    vec bool
}

type _RegFile interface {
    Dump(indent int) []_Reg
    Compare(other _RegFile, indent int) []_Reg
}

var (
    _regs _RegFile
)
