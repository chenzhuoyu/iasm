package x86_64

import (
    `math`
    `reflect`
)

// RelativeOffset represents an RIP-relative offset.
type RelativeOffset int32

// RoundingControl represents a floating-point rounding option.
type RoundingControl uint8

const (
    // RN_SAE represents "Round Nearest", which is the default rounding option.
    RN_SAE RoundingControl = iota

    // RD_SAE represents "Round Down".
    RD_SAE

    // RU_SAE represents "Round Up".
    RU_SAE

    // RZ_SAE represents "Round towards Zero".
    RZ_SAE
)

var _RC_NAMES = map[RoundingControl]string {
    RN_SAE: "rn-sae",
    RD_SAE: "rd-sae",
    RU_SAE: "ru-sae",
    RZ_SAE: "rz-sae",
}

func (self RoundingControl) String() string {
    if v, ok := _RC_NAMES[self]; ok {
        return v
    } else {
        panic("invalid RoundingControl value")
    }
}

// ExceptionControl represents the "Suppress All Exceptions" flag.
type ExceptionControl uint8

const (
    // SAE represents the flag "Suppress All Exceptions" for floating point operations.
    SAE ExceptionControl = iota
)

func (ExceptionControl) String() string {
    return "sae"
}

// AddressType indicates which kind of value that an Addressable contains.
type AddressType uint

const (
    // None indicates the Addressable does not contain any addressable value.
    None AddressType = iota

    // Memory indicates the Addressable contains a memory address.
    Memory

    // Offset indicates the Addressable contains an RIP-relative offset.
    Offset
)

// Addressable is an union to represent something is memory-addressable.
type Addressable struct {
    Type   AddressType
    Memory MemoryAddress
    Offset RelativeOffset
}

// MemoryOperand represents a memory operand for an instruction.
type MemoryOperand struct {
    Size      int
    Addr      Addressable
    Mask      RegisterMask
    Masked    bool
    Broadcast uint8
}

const (
    _Sizes = 0b10000000100010111 // bit-mask for valid sizes (0, 1, 2, 4, 8, 16)
)

func (self *MemoryOperand) isVMX(evex bool) bool {
    return self.Addr.Type == Memory && self.Addr.Memory.isVMX(evex)
}

func (self *MemoryOperand) isVMY(evex bool) bool {
    return self.Addr.Type == Memory && self.Addr.Memory.isVMY(evex)
}

func (self *MemoryOperand) isVMZ() bool {
    return self.Addr.Type == Memory && self.Addr.Memory.isVMZ()
}

func (self *MemoryOperand) isMem() bool {
    if (_Sizes & (1 << self.Broadcast)) == 0 {
        return false
    } else {
        return self.Addr.Type == Offset || (self.Addr.Type == Memory && self.Addr.Memory.isMem())
    }
}

func (self *MemoryOperand) isSize(n int) bool {
    return self.Size == 0 || self.Size == n
}

func (self *MemoryOperand) isBroadcast(n int, b uint8) bool {
    return self.Size == n && self.Broadcast == b
}

func (self *MemoryOperand) ensureAddrValid() {
    switch self.Addr.Type {
        case None   : break
        case Offset : break
        case Memory : self.Addr.Memory.EnsureValid()
        default     : panic("invalid address type")
    }
}

func (self *MemoryOperand) ensureSizeValid() {
    if (_Sizes & (1 << self.Size)) == 0 {
        panic("invalid memory operand size")
    }
}

func (self *MemoryOperand) ensureBroadcastValid() {
    if (_Sizes & (1 << self.Broadcast)) == 0 {
        panic("invalid memory operand broadcast")
    }
}

// EnsureValid checks if the memory operand is valid, if not, it panics.
func (self *MemoryOperand) EnsureValid() {
    self.ensureAddrValid()
    self.ensureSizeValid()
    self.ensureBroadcastValid()
}

// MemoryAddress represents a memory address.
type MemoryAddress struct {
    Base         Register
    Index        Register
    Scale        uint8
    Displacement int32
}

const (
	_Scales = 0b100010111   // bit-mask for valid scales (0, 1, 2, 4, 8)
)

func (self *MemoryAddress) isVMX(evex bool) bool {
    return self.isMemBase() && (self.Index == nil || isXMM(self.Index) || (evex && isEVEXXMM(self.Index)))
}

func (self *MemoryAddress) isVMY(evex bool) bool {
    return self.isMemBase() && (self.Index == nil || isYMM(self.Index) || (evex && isEVEXYMM(self.Index)))
}

func (self *MemoryAddress) isVMZ() bool {
    return self.isMemBase() && (self.Index == nil || isZMM(self.Index))
}

func (self *MemoryAddress) isMem() bool {
    return self.isMemBase() && (self.Index == nil || isReg64(self.Index))
}

func (self *MemoryAddress) isMemBase() bool {
    return (self.Base != nil || self.Index != nil) &&   // must have at least one of `Base` or `Index`
           (self.Base == nil || isReg64(self.Base)) &&  // `Base` must be 64-bit if present
           (self.Scale == 0) == (self.Index == nil) &&  // `Scale` and `Index` depends on each other
           (_Scales & (1 << self.Scale)) != 0           // `Scale` can only be 0, 1, 2, 4 or 8
}

// EnsureValid checks if the memory address is valid, if not, it panics.
func (self *MemoryAddress) EnsureValid() {
    if !self.isMemBase() || (self.Index != nil && !isIndexable(self.Index)) {
        panic("not a valid memory address")
    }
}

// Ptr constructs a simple memory operand with base and displacement.
func Ptr(base Register, disp int32) *MemoryOperand {
    return Sib(base, nil, 0, disp)
}

// Sib constructs a simple memory operand that represents a complete memory address.
func Sib(base Register, index Register, scale uint8, disp int32) (v *MemoryOperand) {
    v = newMemoryOperand()
    v.Addr.Type = Memory
    v.Addr.Memory.Base = base
    v.Addr.Memory.Index = index
    v.Addr.Memory.Scale = scale
    v.Addr.Memory.Displacement = disp
    v.EnsureValid()
    return
}

/** Operand Matching Helpers **/

func isInt(k reflect.Kind) bool {
    switch k {
        case reflect.Int   : fallthrough
        case reflect.Int8  : fallthrough
        case reflect.Int16 : fallthrough
        case reflect.Int32 : fallthrough
        case reflect.Int64 : return true
        default            : return false
    }
}

func isUint(k reflect.Kind) bool {
    switch k {
        case reflect.Uint    : fallthrough
        case reflect.Uint8   : fallthrough
        case reflect.Uint16  : fallthrough
        case reflect.Uint32  : fallthrough
        case reflect.Uint64  : fallthrough
        case reflect.Uintptr : return true
        default              : return false
    }
}

func asInt64(v interface{}) (int64, bool) {
    if isSpecial(v) {
        return 0, false
    } else if x := reflect.ValueOf(v); isInt(x.Kind()) {
        return x.Int(), true
    } else if isUint(x.Kind()) {
        return int64(x.Uint()), true
    } else {
        return 0, false
    }
}

func inRange(v interface{}, low int64, high int64) bool {
    x, ok := asInt64(v)
    return ok && x >= low && x <= high
}

func isSpecial(v interface{}) bool {
    switch v.(type) {
        case Register8        : return true
        case Register16       : return true
        case Register32       : return true
        case Register64       : return true
        case KRegister        : return true
        case MMRegister       : return true
        case XMMRegister      : return true
        case YMMRegister      : return true
        case ZMMRegister      : return true
        case RelativeOffset   : return true
        case RoundingControl  : return true
        case ExceptionControl : return true
        default               : return false
    }
}

func isIndexable(v interface{}) bool {
    return isZMM(v) || isReg64(v) || isEVEXXMM(v) || isEVEXYMM(v)
}

func isAccumulator(v interface{}) bool {
    switch r := v.(type) {
        case Register8  : return r == AL
        case Register16 : return r == AX
        case Register32 : return r == EAX
        case Register64 : return r == RAX
        default         : return false
    }
}

func isImm4        (v interface{}) bool { return inRange(v, 0, 15) }
func isImm8        (v interface{}) bool { return inRange(v, math.MinInt8, math.MaxUint8) }
func isImm16       (v interface{}) bool { return inRange(v, math.MinInt16, math.MaxUint16) }
func isImm32       (v interface{}) bool { return inRange(v, math.MinInt32, math.MaxUint32) }
func isImm64       (v interface{}) bool { _, r := asInt64(v)           ; return r }
func isConst1      (v interface{}) bool { x, r := asInt64(v)           ; return r && x == 1 }
func isConst3      (v interface{}) bool { x, r := asInt64(v)           ; return r && x == 3 }
func isRel8        (v interface{}) bool { x, r := v.(RelativeOffset)   ; return r && x >= math.MinInt8 && x <= math.MaxInt8 }
func isRel32       (v interface{}) bool { _, r := v.(RelativeOffset)   ; return r }
func isReg8        (v interface{}) bool { _, r := v.(Register8)        ; return r }
func isReg8REX     (v interface{}) bool { x, r := v.(Register8)        ; return r && (x & 0x80) == 0 && x >= SPL }
func isReg16       (v interface{}) bool { _, r := v.(Register16)       ; return r }
func isReg32       (v interface{}) bool { _, r := v.(Register32)       ; return r }
func isReg64       (v interface{}) bool { _, r := v.(Register64)       ; return r }
func isMM          (v interface{}) bool { _, r := v.(MMRegister)       ; return r }
func isXMM         (v interface{}) bool { x, r := v.(XMMRegister)      ; return r && x <= XMM15 }
func isEVEXXMM     (v interface{}) bool { _, r := v.(XMMRegister)      ; return r }
func isXMMk        (v interface{}) bool { x, r := v.(MaskedRegister)   ; return isXMM(v) || (r && isXMM(x.Reg) && !x.Mask.Z) }
func isXMMkz       (v interface{}) bool { x, r := v.(MaskedRegister)   ; return isXMM(v) || (r && isXMM(x.Reg)) }
func isYMM         (v interface{}) bool { x, r := v.(YMMRegister)      ; return r && x <= YMM15 }
func isEVEXYMM     (v interface{}) bool { _, r := v.(YMMRegister)      ; return r }
func isYMMk        (v interface{}) bool { x, r := v.(MaskedRegister)   ; return isYMM(v) || (r && isYMM(x.Reg) && !x.Mask.Z) }
func isYMMkz       (v interface{}) bool { x, r := v.(MaskedRegister)   ; return isYMM(v) || (r && isYMM(x.Reg)) }
func isZMM         (v interface{}) bool { _, r := v.(ZMMRegister)      ; return r }
func isZMMk        (v interface{}) bool { x, r := v.(MaskedRegister)   ; return isZMM(v) || (r && isZMM(x.Reg) && !x.Mask.Z) }
func isZMMkz       (v interface{}) bool { x, r := v.(MaskedRegister)   ; return isZMM(v) || (r && isZMM(x.Reg)) }
func isK           (v interface{}) bool { _, r := v.(KRegister)        ; return r }
func isKk          (v interface{}) bool { x, r := v.(MaskedRegister)   ; return isK(v) || (r && isK(x.Reg) && !x.Mask.Z) }
func isM           (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isMem() && x.Broadcast == 0 && !x.Masked }
func isMk          (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isMem() && x.Broadcast == 0 && !(x.Masked && x.Mask.Z) }
func isMkz         (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isMem() && x.Broadcast == 0 }
func isM8          (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isM(v)   && x.isSize(1)  }
func isM16         (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isM(v)   && x.isSize(2)  }
func isM16kz       (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMkz(v) && x.isSize(2)  }
func isM32         (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isM(v)   && x.isSize(4)  }
func isM32k        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMk(v)  && x.isSize(4)  }
func isM32kz       (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMkz(v) && x.isSize(4)  }
func isM64         (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isM(v)   && x.isSize(8)  }
func isM64k        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMk(v)  && x.isSize(8)  }
func isM64kz       (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMkz(v) && x.isSize(8)  }
func isM128        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isM(v)   && x.isSize(16) }
func isM128kz      (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMkz(v) && x.isSize(16) }
func isM256        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isM(v)   && x.isSize(32) }
func isM256kz      (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMkz(v) && x.isSize(32) }
func isM512        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isM(v)   && x.isSize(64) }
func isM512kz      (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && isMkz(v) && x.isSize(64) }
func isM64M32bcst  (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return isM64(v)  || (r && x.isBroadcast(4, 2)) }
func isM128M32bcst (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return isM128(v) || (r && x.isBroadcast(4, 4)) }
func isM256M32bcst (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return isM256(v) || (r && x.isBroadcast(4, 8)) }
func isM512M32bcst (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return isM512(v) || (r && x.isBroadcast(4, 16)) }
func isM128M64bcst (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return isM128(v) || (r && x.isBroadcast(8, 2)) }
func isM256M64bcst (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return isM256(v) || (r && x.isBroadcast(8, 4)) }
func isM512M64bcst (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return isM512(v) || (r && x.isBroadcast(8, 8)) }
func isVMX         (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMX(false) && !x.Masked }
func isEVEXVMX     (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMX(true) && !x.Masked }
func isVMXk        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMX(true) }
func isVMY         (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMY(false) && !x.Masked }
func isEVEXVMY     (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMY(true) && !x.Masked }
func isVMYk        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMY(true) }
func isVMZ         (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMZ() && !x.Masked }
func isVMZk        (v interface{}) bool { x, r := v.(*MemoryOperand)   ; return r && x.isVMZ() }
func isSAE         (v interface{}) bool { _, r := v.(ExceptionControl) ; return r }
func isER          (v interface{}) bool { _, r := v.(RoundingControl)  ; return r }

func isImmExt(v interface{}, ext int, min int64, max int64) bool {
    if x, ok := asInt64(v); !ok {
        return false
    } else if m := int64(1) << (8 * ext); x < m && x >= m + min {
        return true
    } else {
        return x <= max && x >= min
    }
}

func isImm8Ext(v interface{}, ext int) bool {
    return isImmExt(v, ext, math.MinInt8, math.MaxInt8)
}

func isImm32Ext(v interface{}, ext int) bool {
    return isImmExt(v, ext, math.MinInt32, math.MaxInt32)
}
