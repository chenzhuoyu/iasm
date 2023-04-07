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
