package aarch64

import (
    `github.com/chenzhuoyu/iasm/asm`
)

func mext   (v interface{}) asm.MemoryAddressExtension { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Ext }
func mbase  (v interface{}) asm.Register               { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Base }
func midx   (v interface{}) asm.Register               { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Index }
func moffs  (v interface{}) int32                      { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Offset }
func velm   (v interface{}) VecFormat                  { return v.(Vector).Format() }
func vfmt   (v interface{}) VecFormat                  { return v.(VRegister).Format() }
func vidxr  (v interface{}) uint8                      { return v.(VidxRegister).Index() }
func vmoder (v interface{}) VecIndexMode               { return v.(VidxRegister).IndexMode() }
func vmodei (v interface{}) VecIndexMode               { return v.(IndexedVector).IndexMode() }

func modn(v interface{}) uint8 {
    if m, ok := v.(Modifier); !ok {
        return 0
    } else {
        return m.Amount()
    }
}

func modt(v interface{}) ModType {
    if m, ok := v.(Modifier); !ok {
        return ModInvalid
    } else {
        return m.Type()
    }
}

func abs12(lb *asm.Label) uint32 {
    return uint32(lb.Address() & 0xfff)
}

func rel14(lb *asm.Label, pc uintptr) uint32 {
    if d := lb.Address() - pc; d & 0b11 != 0 {
        panic("aarch64: target address is not aligned")
    } else if r := uint32(int64(d) >> 2); r &^ 0x3fff != 0 {
        panic("aarch64: target is too far to fit into 14 bit relative address")
    } else {
        return r
    }
}

func rel19(lb *asm.Label, pc uintptr) uint32 {
    if d := lb.Address() - pc; d & 0b11 != 0 {
        panic("aarch64: target address is not aligned")
    } else if r := uint32(int64(d) >> 2); r &^ 0x7ffff != 0 {
        panic("aarch64: target is too far to fit into 19 bit relative address")
    } else {
        return r
    }
}

func rel26(lb *asm.Label, pc uintptr) uint32 {
    if d := lb.Address() - pc; d & 0b11 != 0 {
        panic("aarch64: target address is not aligned")
    } else if r := uint32(int64(d) >> 2); r &^ 0x3ffffff != 0 {
        panic("aarch64: target is too far to fit into 26 bit relative address")
    } else {
        return r
    }
}

func reladr(lb *asm.Label, pc uintptr, adrp bool) uint32 {
    base := pc
    dest := lb.Address()

    /* align to page for ADRP */
    if adrp {
        base >>= 12
        dest >>= 12
    }

    /* calculate the relative address */
    if rel := dest - base; rel &^ 0x1fffff != 0 {
        panic("aarch64: target is too far to fit into 21 bit relative address")
    } else {
        return uint32(rel)
    }
}
