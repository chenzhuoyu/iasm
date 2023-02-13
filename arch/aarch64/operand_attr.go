package aarch64

import (
    `github.com/chenzhuoyu/iasm/asm`
)

func mext   (v interface{}) asm.MemoryAddressExtension { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Ext }
func mbase  (v interface{}) asm.Register               { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Base }
func midx   (v interface{}) asm.Register               { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Index }
func moffs  (v interface{}) int32                      { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Offset }
func vfmt   (v interface{}) SIMDVectorArrangement      { return v.(SIMDRegister128r).Arrangement() }

func modn(v interface{}) uint8 {
    if m, ok := v.(Modifier); !ok {
        return 0
    } else {
        return m.Amount()
    }
}
