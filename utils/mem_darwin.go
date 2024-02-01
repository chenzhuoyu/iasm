package utils

import (
    `unsafe`
)

// #include <pthread.h>
// #include <libkern/OSCacheControl.h>
import "C"

func EnableMemWrite(dest *byte, size int) {
    C.pthread_jit_write_protect_np(0)
    C.sys_icache_invalidate(unsafe.Pointer(dest), C.ulong(size))
}

func DisableMemWrite(dest *byte, size int) {
    C.pthread_jit_write_protect_np(1)
    C.sys_icache_invalidate(unsafe.Pointer(dest), C.ulong(size))
}
