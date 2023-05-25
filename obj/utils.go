package obj

import (
    `os`
    `os/exec`
)

const (
    defaultLd = "/usr/bin/ld"
)

func min(a, b int) int {
    if a < b {
        return a
    } else {
        return b
    }
}

func str16b(s string) (v [16]byte) {
    copy(v[:], s[:min(len(s), len(v) - 1)])
    return
}

func findldcmd() string {
    if cmd := os.Getenv("LD"); cmd != "" {
        return cmd
    } else if ret, err := exec.LookPath("ld"); err == nil {
        return ret
    } else {
        return defaultLd
    }
}
