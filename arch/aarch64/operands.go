package aarch64

import (
    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// IndexMode stores the aarch64 specific extensions of asm.MemoryAddress
type IndexMode uint8

const (
    NoIndex IndexMode = iota    // Unsigned-offset form "[<Xn|SP>, #<pimm>]"
    PreIndex                    // Pre-index form "[<Xn|SP>, #<simm>]!"
    PostIndex                   // Post-indexing form "[<Xn|SP>], #<simm>"
)

func (IndexMode) Free()                   {}
func (IndexMode) Sealed(tag.Tag)          {}
func (IndexMode) MemoryAddressExtension() {}

func (self IndexMode) String(addr asm.MemoryAddress) string {
    // TODO implement me
    panic("implement me")
}

func (self IndexMode) EnsureValid(addr asm.MemoryAddress) {
    // TODO implement me
    panic("implement me")
}

// ShiftType represents one of the register shift types.
type ShiftType uint8

const (
    LSL ShiftType = 0b00    // Logical shift left
    MSL ShiftType = 0b00    // Masking shift left
    LSR ShiftType = 0b01    // Logical shift right
    ASR ShiftType = 0b10    // Arithmetic shift right
)

func (ShiftType) Free()                   {}
func (ShiftType) Sealed(tag.Tag)          {}
func (ShiftType) MemoryAddressExtension() {}

func (self ShiftType) String(addr asm.MemoryAddress) string {
    // TODO implement me
    panic("implement me")
}

func (self ShiftType) EnsureValid(addr asm.MemoryAddress) {
    // TODO implement me
    panic("implement me")
}

// Extension represents one of the register extensions.
type Extension uint8

const (
    UXTB Extension = 0b000  // Unsigned extension from byte to 64-bit
    UXTH Extension = 0b001  // Unsigned extension from half-word to 64-bit
    UXTW Extension = 0b010  // Unsigned extension from word to 64-bit
    UXTX Extension = 0b011  // Unsigned extension from 64-bit to 64-bit
    SXTB Extension = 0b100  // Signed extension from byte to 64-bit
    SXTH Extension = 0b101  // Signed extension from half-word to 64-bit
    SXTW Extension = 0b110  // Signed extension from word to 64-bit
    SXTX Extension = 0b111  // Signed extension from 64-bit to 64-bit
)

func (Extension) Free()                   {}
func (Extension) Sealed(tag.Tag)          {}
func (Extension) MemoryAddressExtension() {}

func (self Extension) String(addr asm.MemoryAddress) string {
    // TODO implement me
    panic("implement me")
}

func (self Extension) EnsureValid(addr asm.MemoryAddress) {
    // TODO implement me
    panic("implement me")
}

type (
    _MemOpExt struct{}
)

// MemoryOperandExtension is the aarch specific extensions of asm.MemoryOperand
var MemoryOperandExtension _MemOpExt

func (_MemOpExt) Free()                   {}
func (_MemOpExt) Sealed(tag.Tag)          {}
func (_MemOpExt) MemoryOperandExtension() {}

func (_MemOpExt) String(mem *asm.MemoryOperand) string {
    return mem.Addr.String()
}

func (_MemOpExt) EnsureValid(mem *asm.MemoryOperand) {
    mem.Addr.EnsureValid()
}

// Mem constructs a memory operand from an addressable value.
func Mem(addr asm.Addressable) (v *asm.MemoryOperand) {
    v = asm.CreateMemoryOperand(MemoryOperandExtension)
    v.Addr = addr
    v.EnsureValid()
    return
}

// Offs constructs a simple memory operand with base and offset.
func Offs(base Register64, offs int32, mode IndexMode) *asm.MemoryOperand {
    return Mem(asm.MemoryAddress {
        Ext    : mode,
        Base   : base,
        Offset : offs,
    })
}
