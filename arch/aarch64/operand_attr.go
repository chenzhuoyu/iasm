package aarch64

import (
    `fmt`

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

func adradj(ra uintptr, adrp bool) uintptr {
    if !adrp {
        return ra
    } else {
        return ra >> 12
    }
}

func adrrel(ra asm.RelativeOffset, adrp bool) uint32 {
    if rel := adradj(uintptr(ra), adrp); rel >> 21 != 0 && rel >> 21 != 0x7ffffffffff {
        panic("aarch64: target is too far to fit into 21 bit relative address")
    } else {
        return uint32(rel)
    }
}

func adroffs(lb *asm.Label, pc uintptr, adrp bool) uint32 {
    if rel := adradj(lb.Address(), adrp) - adradj(pc, adrp); rel >> 21 != 0 && rel >> 21 != 0x7ffffffffff {
        panic("aarch64: target is too far to fit into 21 bit relative address")
    } else {
        return uint32(rel)
    }
}

func rel14(ra asm.RelativeOffset) uint32 { return pcreloffs(uintptr(ra), 14) }
func rel19(ra asm.RelativeOffset) uint32 { return pcreloffs(uintptr(ra), 19) }
func rel26(ra asm.RelativeOffset) uint32 { return pcreloffs(uintptr(ra), 26) }

func pcrel14(lb *asm.Label, pc uintptr) uint32 { return pcreloffs(lb.Address() - pc, 14) }
func pcrel19(lb *asm.Label, pc uintptr) uint32 { return pcreloffs(lb.Address() - pc, 19) }
func pcrel26(lb *asm.Label, pc uintptr) uint32 { return pcreloffs(lb.Address() - pc, 26) }

func pcreloffs(offs uintptr, bits int) uint32 {
    if offs & 0b11 != 0 {
        panic("aarch64: target address is not aligned")
    } else if r := uintptr(int64(offs) >> 2); r >> bits != 0 && r >> bits != (1 << (64 - bits)) - 1 {
        panic(fmt.Sprintf("aarch64: target is too far to fit into %d bit relative address", bits))
    } else {
        return uint32(r)
    }
}
