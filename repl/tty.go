package repl

import (
    `syscall`
    `unsafe`
)

func isatty(fd uintptr) bool {
    var v syscall.Termios
    _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlCode, uintptr(unsafe.Pointer(&v)), 0, 0, 0)
    return err == 0
}