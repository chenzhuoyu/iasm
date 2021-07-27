package repl

type _ExecutorAMD64 struct {
    after  _RegFileAMD64
    before _RegFileAMD64
}

//go:nosplit
//go:noescape
//goland:noinspection GoUnusedParameter
func execaddr(addr uintptr, before *_RegFileAMD64, after *_RegFileAMD64)

func (self *_ExecutorAMD64) Execute(addr uintptr) (_RegFile, _RegFile, error) {
    execaddr(addr, &self.before, &self.after)
    return &self.before, &self.after, nil
}

func init() {
    _exec = new(_ExecutorAMD64)
}
