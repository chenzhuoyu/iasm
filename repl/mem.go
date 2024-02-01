package repl

import (
    `os`
    `syscall`
    `unsafe`
)

type _Slice struct {
    ptr uintptr
    len int
    cap int
}

type _Memory struct {
    size uint64
    addr uintptr
}

func (self _Memory) p() (v unsafe.Pointer) {
    *(*uintptr)(unsafe.Pointer(&v)) = self.addr
    return
}

func (self _Memory) buf() (v []byte) {
    p := (*_Slice)(unsafe.Pointer(&v))
    p.ptr = self.addr
    p.len = int(self.size)
    p.cap = int(self.size)
    return
}

const (
    _RWX = syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
)

var (
    _PageSize = uint64(os.Getpagesize())
)

func mmap(nb uint64) (_Memory, error) {
    nb = (((nb - 1) / _PageSize) + 1) * _PageSize
    mm, _, err := syscall.RawSyscall6(syscall.SYS_MMAP, 0, uintptr(nb), _RWX, _MMAP_FLAGS, 0, 0)

    /* check for errors */
    if err != 0 {
        return _Memory{}, err
    } else {
        return _Memory { addr: mm, size: nb }, nil
    }
}

func munmap(m _Memory) error {
    if _, _, err := syscall.RawSyscall(syscall.SYS_MUNMAP, m.addr, uintptr(m.size), 0); err != 0 {
        return err
    } else {
        return nil
    }
}
