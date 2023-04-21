package x86_64

import (
    `fmt`
    `math`
    `reflect`
    `strconv`
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/rt`
)

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

const (
    _Sizes  = 0b10000000100010111   // bit-mask for valid sizes (0, 1, 2, 4, 8, 16)
    _Scales = 0b100010111           // bit-mask for valid scales (0, 1, 2, 4, 8)
)

// Scale is the x86_64 specific extensions of asm.MemoryAddress (e.g. the scale value).
type Scale uint8

func (Scale) Free()                   {}
func (Scale) MemoryAddressExtension() {}

func (self Scale) isVMX(addr asm.MemoryAddress, evex bool) bool {
    return self.isMemBase(addr) && (addr.Index == nil || isXMM(addr.Index) || (evex && isEVEXXMM(addr.Index)))
}

func (self Scale) isVMY(addr asm.MemoryAddress, evex bool) bool {
    return self.isMemBase(addr) && (addr.Index == nil || isYMM(addr.Index) || (evex && isEVEXYMM(addr.Index)))
}

func (self Scale) isVMZ(addr asm.MemoryAddress) bool {
    return self.isMemBase(addr) && (addr.Index == nil || isZMM(addr.Index))
}

func (self Scale) isMem(addr asm.MemoryAddress) bool {
    return self.isMemBase(addr) && (addr.Index == nil || isReg64(addr.Index))
}

func (self Scale) isMemBase(addr asm.MemoryAddress) bool {
    return (addr.Base == nil || isReg64(addr.Base)) &&  // `Base` must be 64-bit if present
           (self == 0) == (addr.Index == nil) &&        // `Scale` and `Index` depends on each other
           (_Scales & (1 << self)) != 0                 // `Scale` can only be 0, 1, 2, 4 or 8
}

func (self Scale) String() string {
    return fmt.Sprintf("Scale(%d)", uint8(self))
}

func (self Scale) Display(addr asm.MemoryAddress) string {
    var dp int
    var sb strings.Builder

    /* the displacement part */
    if dp = int(addr.Offset); dp != 0 {
        sb.WriteString(strconv.Itoa(dp))
    }

    /* the base register */
    if sb.WriteByte('('); addr.Base != nil {
        sb.WriteByte('%')
        sb.WriteString(addr.Base.String())
    }

    /* index is optional */
    if addr.Index != nil {
        sb.WriteString(",%")
        sb.WriteString(addr.Index.String())

        /* scale is also optional */
        if self >= 2 {
            sb.WriteByte(',')
            sb.WriteString(strconv.Itoa(int(self)))
        }
    }

    /* close the bracket */
    sb.WriteByte(')')
    return sb.String()
}

func (self Scale) EnsureValid(addr asm.MemoryAddress) {
    if !self.isMemBase(addr) || (addr.Index != nil && !isIndexable(addr.Index)) {
        panic("not a valid memory address")
    }
}

// MemoryAddress constructs an x86_64 specific asm.MemoryAddress
func MemoryAddress(base asm.Register, index asm.Register, scale uint8, disp int32) asm.MemoryAddress {
    return asm.MemoryAddress {
        Ext    : Scale(scale),
        Base   : base,
        Index  : index,
        Offset : disp,
    }
}

// MemoryOperandExtension stores the x86_64 specific extensions of asm.MemoryOperand
type MemoryOperandExtension struct {
    Size      int
    Mask      RegisterMask
    Masked    bool
    Broadcast uint8
}

func (self *MemoryOperandExtension) isVMX(mem *asm.MemoryOperand, evex bool) bool {
    if v, ok := mem.Addr.(asm.MemoryAddress); !ok {
        return false
    } else {
        return v.Ext.(Scale).isVMX(v, evex)
    }
}

func (self *MemoryOperandExtension) isVMY(mem *asm.MemoryOperand, evex bool) bool {
    if v, ok := mem.Addr.(asm.MemoryAddress); !ok {
        return false
    } else {
        return v.Ext.(Scale).isVMY(v, evex)
    }
}

func (self *MemoryOperandExtension) isVMZ(mem *asm.MemoryOperand) bool {
    if v, ok := mem.Addr.(asm.MemoryAddress); !ok {
        return false
    } else {
        return v.Ext.(Scale).isVMZ(v)
    }
}

func (self *MemoryOperandExtension) isMem(mem *asm.MemoryOperand) bool {
    if (_Sizes & (1 << self.Broadcast)) == 0 {
        return false
    } else {
        switch v := mem.Addr.(type) {
            case *asm.Label         : return true
            case asm.MemoryAddress  : return v.Ext.(Scale).isMem(v)
            case asm.RelativeOffset : return true
            default                 : return false
        }
    }
}

func (self *MemoryOperandExtension) isSize(n int) bool {
    return self.Size == 0 || self.Size == n
}

func (self *MemoryOperandExtension) isBroadcast(n int, b uint8) bool {
    return self.Size == n && self.Broadcast == b
}

func (self *MemoryOperandExtension) formatMask() string {
    if !self.Masked {
        return ""
    } else {
        return self.Mask.String()
    }
}

func (self *MemoryOperandExtension) formatBroadcast() string {
    if self.Broadcast == 0 {
        return ""
    } else {
        return fmt.Sprintf("{1to%d}", self.Broadcast)
    }
}

func (self *MemoryOperandExtension) ensureSizeValid() {
    if (_Sizes & (1 << self.Size)) == 0 {
        panic("invalid memory operand size")
    }
}

func (self *MemoryOperandExtension) ensureAddrValid(mem *asm.MemoryOperand) {
    if mem.Addr != nil {
        mem.Addr.EnsureValid()
    }
}

func (self *MemoryOperandExtension) ensureBroadcastValid() {
    if (_Sizes & (1 << self.Broadcast)) == 0 {
        panic("invalid memory operand broadcast")
    }
}

func (self *MemoryOperandExtension) Free()                   { freeMemoryOperandExtension(self) }
func (self *MemoryOperandExtension) MemoryOperandExtension() {}

func (self *MemoryOperandExtension) String() string {
    return fmt.Sprintf("MemOpExt.x86_64(%s:%s)", self.formatMask(), self.formatBroadcast())
}

func (self *MemoryOperandExtension) Display(mem *asm.MemoryOperand) string {
    return mem.Addr.String() + self.formatMask() + self.formatBroadcast()
}

func (self *MemoryOperandExtension) EnsureValid(mem *asm.MemoryOperand) {
    self.ensureSizeValid()
    self.ensureAddrValid(mem)
    self.ensureBroadcastValid()
}

// Abs constructs an absolute addressing memory operand.
func Abs(disp int32) *asm.MemoryOperand {
    return Sib(nil, nil, 0, disp)
}

// Ptr constructs a simple memory operand with base and displacement.
func Ptr(base asm.Register, disp int32) *asm.MemoryOperand {
    return Sib(base, nil, 0, disp)
}

// Sib constructs a simple memory operand that represents a complete memory address.
func Sib(base asm.Register, index asm.Register, scale uint8, disp int32) *asm.MemoryOperand {
    return Mem(MemoryAddress(base, index, scale, disp))
}

// Mem constructs a memory operand from an addressable value.
func Mem(addr asm.Addressable) (v *asm.MemoryOperand) {
    v = asm.CreateMemoryOperand(newMemoryOperandExtension())
    v.Addr = addr
    v.EnsureValid()
    return
}

/** Operand Matching Helpers **/

const _IntMask =
    (1 << reflect.Int    ) |
    (1 << reflect.Int8   ) |
    (1 << reflect.Int16  ) |
    (1 << reflect.Int32  ) |
    (1 << reflect.Int64  ) |
    (1 << reflect.Uint   ) |
    (1 << reflect.Uint8  ) |
    (1 << reflect.Uint16 ) |
    (1 << reflect.Uint32 ) |
    (1 << reflect.Uint64 ) |
    (1 << reflect.Uintptr)

func isInt(k reflect.Kind) bool {
    return (_IntMask & (1 << k)) != 0
}

func asInt64(v interface{}) (int64, bool) {
    if isSpecial(v) {
        return 0, false
    } else if x := rt.AsEface(v); isInt(x.Kind()) {
        return x.ToInt64(), true
    } else {
        return 0, false
    }
}

func asMemOp(v interface{}) (*asm.MemoryOperand, *MemoryOperandExtension, bool) {
    if x, ok := v.(*asm.MemoryOperand); !ok {
        return nil, nil, false
    } else {
        return x, x.Ext.(*MemoryOperandExtension), true
    }
}

func isSpecial(v interface{}) bool {
    switch v.(type) {
        case Register8          : return true
        case Register16         : return true
        case Register32         : return true
        case Register64         : return true
        case KRegister          : return true
        case MMRegister         : return true
        case XMMRegister        : return true
        case YMMRegister        : return true
        case ZMMRegister        : return true
        case RoundingControl    : return true
        case ExceptionControl   : return true
        case asm.RelativeOffset : return true
        default                 : return false
    }
}

func isIndexable(v interface{}) bool {
    return isZMM(v) || isReg64(v) || isEVEXXMM(v) || isEVEXYMM(v)
}

func isImm4   (v interface{}) bool { x, r := asInt64(v); return r && x >= 0 && x <= 15 }
func isImm8   (v interface{}) bool { x, r := asInt64(v); return r && x >= math.MinInt8 && x <= math.MaxUint8 }
func isImm16  (v interface{}) bool { x, r := asInt64(v); return r && x >= math.MinInt16 && x <= math.MaxUint16 }
func isImm32  (v interface{}) bool { x, r := asInt64(v); return r && x >= math.MinInt32 && x <= math.MaxUint32 }
func isImm64  (v interface{}) bool { _, r := asInt64(v); return r }
func isConst1 (v interface{}) bool { x, r := asInt64(v); return r && x == 1 }
func isConst3 (v interface{}) bool { x, r := asInt64(v); return r && x == 3 }

func isRel8    (v interface{}) bool { x, r := v.(asm.RelativeOffset) ; return r && x >= math.MinInt8 && x <= math.MaxInt8 }
func isRel32   (v interface{}) bool { _, r := v.(asm.RelativeOffset) ; return r }
func isLabel   (v interface{}) bool { _, r := v.(*asm.Label)         ; return r }
func isReg8    (v interface{}) bool { _, r := v.(Register8)          ; return r }
func isReg8REX (v interface{}) bool { x, r := v.(Register8)          ; return r && (x & 0x80) == 0 && x >= SPL }
func isReg16   (v interface{}) bool { _, r := v.(Register16)         ; return r }
func isReg32   (v interface{}) bool { _, r := v.(Register32)         ; return r }
func isReg64   (v interface{}) bool { _, r := v.(Register64)         ; return r }
func isMM      (v interface{}) bool { _, r := v.(MMRegister)         ; return r }
func isXMM     (v interface{}) bool { x, r := v.(XMMRegister)        ; return r && x <= XMM15 }
func isEVEXXMM (v interface{}) bool { _, r := v.(XMMRegister)        ; return r }
func isXMMk    (v interface{}) bool { x, r := v.(MaskedRegister)     ; return isXMM(v) || (r && isXMM(x.Reg) && !x.Mask.Z) }
func isXMMkz   (v interface{}) bool { x, r := v.(MaskedRegister)     ; return isXMM(v) || (r && isXMM(x.Reg)) }
func isYMM     (v interface{}) bool { x, r := v.(YMMRegister)        ; return r && x <= YMM15 }
func isEVEXYMM (v interface{}) bool { _, r := v.(YMMRegister)        ; return r }
func isYMMk    (v interface{}) bool { x, r := v.(MaskedRegister)     ; return isYMM(v) || (r && isYMM(x.Reg) && !x.Mask.Z) }
func isYMMkz   (v interface{}) bool { x, r := v.(MaskedRegister)     ; return isYMM(v) || (r && isYMM(x.Reg)) }
func isZMM     (v interface{}) bool { _, r := v.(ZMMRegister)        ; return r }
func isZMMk    (v interface{}) bool { x, r := v.(MaskedRegister)     ; return isZMM(v) || (r && isZMM(x.Reg) && !x.Mask.Z) }
func isZMMkz   (v interface{}) bool { x, r := v.(MaskedRegister)     ; return isZMM(v) || (r && isZMM(x.Reg)) }
func isK       (v interface{}) bool { _, r := v.(KRegister)          ; return r }
func isKk      (v interface{}) bool { x, r := v.(MaskedRegister)     ; return isK(v) || (r && isK(x.Reg) && !x.Mask.Z) }
func isSAE     (v interface{}) bool { _, r := v.(ExceptionControl)   ; return r }
func isER      (v interface{}) bool { _, r := v.(RoundingControl)    ; return r }

func isM           (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isMem(m) && x.Broadcast == 0 && !x.Masked }
func isMk          (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isMem(m) && x.Broadcast == 0 && !(x.Masked && x.Mask.Z) }
func isMkz         (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isMem(m) && x.Broadcast == 0 }
func isM8          (v interface{}) bool { _, x, r := asMemOp(v); return r && isM(v)   && x.isSize(1) }
func isM16         (v interface{}) bool { _, x, r := asMemOp(v); return r && isM(v)   && x.isSize(2) }
func isM16kz       (v interface{}) bool { _, x, r := asMemOp(v); return r && isMkz(v) && x.isSize(2) }
func isM32         (v interface{}) bool { _, x, r := asMemOp(v); return r && isM(v)   && x.isSize(4) }
func isM32k        (v interface{}) bool { _, x, r := asMemOp(v); return r && isMk(v)  && x.isSize(4) }
func isM32kz       (v interface{}) bool { _, x, r := asMemOp(v); return r && isMkz(v) && x.isSize(4) }
func isM64         (v interface{}) bool { _, x, r := asMemOp(v); return r && isM(v)   && x.isSize(8) }
func isM64k        (v interface{}) bool { _, x, r := asMemOp(v); return r && isMk(v)  && x.isSize(8) }
func isM64kz       (v interface{}) bool { _, x, r := asMemOp(v); return r && isMkz(v) && x.isSize(8) }
func isM128        (v interface{}) bool { _, x, r := asMemOp(v); return r && isM(v)   && x.isSize(16) }
func isM128kz      (v interface{}) bool { _, x, r := asMemOp(v); return r && isMkz(v) && x.isSize(16) }
func isM256        (v interface{}) bool { _, x, r := asMemOp(v); return r && isM(v)   && x.isSize(32) }
func isM256kz      (v interface{}) bool { _, x, r := asMemOp(v); return r && isMkz(v) && x.isSize(32) }
func isM512        (v interface{}) bool { _, x, r := asMemOp(v); return r && isM(v)   && x.isSize(64) }
func isM512kz      (v interface{}) bool { _, x, r := asMemOp(v); return r && isMkz(v) && x.isSize(64) }
func isM64M32bcst  (v interface{}) bool { _, x, r := asMemOp(v); return isM64(v)  || (r && x.isBroadcast(4, 2)) }
func isM128M32bcst (v interface{}) bool { _, x, r := asMemOp(v); return isM128(v) || (r && x.isBroadcast(4, 4)) }
func isM256M32bcst (v interface{}) bool { _, x, r := asMemOp(v); return isM256(v) || (r && x.isBroadcast(4, 8)) }
func isM512M32bcst (v interface{}) bool { _, x, r := asMemOp(v); return isM512(v) || (r && x.isBroadcast(4, 16)) }
func isM128M64bcst (v interface{}) bool { _, x, r := asMemOp(v); return isM128(v) || (r && x.isBroadcast(8, 2)) }
func isM256M64bcst (v interface{}) bool { _, x, r := asMemOp(v); return isM256(v) || (r && x.isBroadcast(8, 4)) }
func isM512M64bcst (v interface{}) bool { _, x, r := asMemOp(v); return isM512(v) || (r && x.isBroadcast(8, 8)) }
func isVMX         (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMX(m, false) && !x.Masked }
func isEVEXVMX     (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMX(m, true)  && !x.Masked }
func isVMXk        (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMX(m, true) }
func isVMY         (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMY(m, false) && !x.Masked }
func isEVEXVMY     (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMY(m, true)  && !x.Masked }
func isVMYk        (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMY(m, true) }
func isVMZ         (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMZ(m) && !x.Masked }
func isVMZk        (v interface{}) bool { m, x, r := asMemOp(v); return r && x.isVMZ(m) }

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
