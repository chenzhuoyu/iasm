package repl

import (
    `runtime`
)

func symbol(addr uint64) string {
    return runtime.FuncForPC(uintptr(addr)).Name()
}
