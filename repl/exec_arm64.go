package repl

type _ExecutorARM64 struct {
    after  _RegFileARM64
    before _RegFileARM64
}

//go:nosplit
//go:noescape
//goland:noinspection GoUnusedParameter
func execaddr(addr uintptr, before *_RegFileARM64, after *_RegFileARM64)

func (self *_ExecutorARM64) Execute(addr uintptr) (_RegFile, _RegFile, error) {
    execaddr(addr, &self.before, &self.after)
    return &self.before, &self.after, nil
}

func init() {
    _exec = new(_ExecutorARM64)
}
