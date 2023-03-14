package aarch64

import (
    `github.com/chenzhuoyu/iasm/asm`
)

func mext   (v interface{}) asm.MemoryAddressExtension { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Ext }
func mbase  (v interface{}) asm.Register               { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Base }
func midx   (v interface{}) asm.Register               { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Index }
func moffs  (v interface{}) int32                      { return v.(*asm.MemoryOperand).Addr.(asm.MemoryAddress).Offset }
func vstrr  (v interface{}) SIMDVectorStructure        { return v.(_Indexed128r).Structure() }
func vstr1  (v interface{}) SIMDVectorStructure        { return v.(_IndexedVec1).Structure() }
func vstr2  (v interface{}) SIMDVectorStructure        { return v.(_IndexedVec2).Structure() }
func vstr3  (v interface{}) SIMDVectorStructure        { return v.(_IndexedVec3).Structure() }
func vstr4  (v interface{}) SIMDVectorStructure        { return v.(_IndexedVec4).Structure() }
func vfmt   (v interface{}) SIMDVectorArrangement      { return v.(SIMDRegister128v).Arrangement() }

func modn(v interface{}) uint8 {
    if m, ok := v.(Modifier); !ok {
        return 0
    } else {
        return m.Amount()
    }
}
