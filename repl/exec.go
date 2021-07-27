package repl

type _Executor interface {
    Execute(addr uintptr) (_RegFile, _RegFile, error)
}

var (
    _exec _Executor
)
