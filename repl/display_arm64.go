package repl

import (
    `fmt`
)

func nzcv(nzcv uint64) string {
    nf := (nzcv >> 31) & 1
    zf := (nzcv >> 30) & 1
    cf := (nzcv >> 29) & 1
    vf := (nzcv >> 28) & 1
    return fmt.Sprintf("%c%c%c%c", "-n"[nf], "-z"[zf], "-c"[cf], "-v"[vf])
}
