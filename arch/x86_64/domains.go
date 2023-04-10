package x86_64

import (
    `github.com/chenzhuoyu/iasm/asm`
)

const (
    DomainMMXSSE = asm.DomainArchSpecific + iota
    DomainAVX
    DomainFMA
    DomainCrypto
    DomainMask
    DomainAMDSpecific
    DomainMisc
)
