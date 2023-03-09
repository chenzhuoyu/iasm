package aarch64

import (
    `math`
    `reflect`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/rt`
)

const _IntMask =
    (1 << reflect.Int  ) |
    (1 << reflect.Int8 ) |
    (1 << reflect.Int16) |
    (1 << reflect.Int32) |
    (1 << reflect.Int64)

const _UintMask =
    (1 << reflect.Uint   ) |
    (1 << reflect.Uint8  ) |
    (1 << reflect.Uint16 ) |
    (1 << reflect.Uint32 ) |
    (1 << reflect.Uint64 ) |
    (1 << reflect.Uintptr)

const _FloatMask =
    (1 << reflect.Float32) |
    (1 << reflect.Float64)

func isInt(k reflect.Kind) bool {
    return (_IntMask & (1 << k)) != 0
}

func isUint(k reflect.Kind) bool {
    return (_IntMask & (1 << k)) != 0
}

func isFloat(k reflect.Kind) bool {
    return (_FloatMask & (1 << k)) != 0
}

func isInt32(v interface{}) bool {
    x, ok := asInt64(v)
    return ok && x >= math.MinInt32 && x <= math.MaxInt32
}

func isSpecial(v interface{}) bool {
    switch v.(type) {
        case Register32         : return true
        case Register64         : return true
        case SIMDRegister8      : return true
        case SIMDRegister16     : return true
        case SIMDRegister32     : return true
        case SIMDRegister64     : return true
        case SIMDRegister128    : return true
        case SIMDRegister128r   : return true
        case SIMDRegister128v   : return true
        case SIMDVector1        : return true
        case SIMDVector2        : return true
        case SIMDVector3        : return true
        case SIMDVector4        : return true
        case _Indexed128r       : return true
        case _IndexedVec1       : return true
        case _IndexedVec2       : return true
        case _IndexedVec3       : return true
        case _IndexedVec4       : return true
        case PStateField: return true
        case SystemRegister     : return true
        case asm.RelativeOffset : return true
        default                 : return false
    }
}

func isSameType(x, y interface{}) bool {
    return rt.TypeOf(x) == rt.TypeOf(y)
}

func isSameSize(x, y interface{}) bool {
    if a, ok := x.(SIMDRegister128v); !ok {
        return false
    } else if b, ok := y.(SIMDRegister128v); !ok {
        return false
    } else {
        return a.Arrangement() == b.Arrangement()
    }
}

func isLabel     (v interface{}) bool { _, f := v.(*asm.Label)       ; return f }
func isXr        (v interface{}) bool { x, f := v.(Register64)       ; return f && x != SP }
func isWr        (v interface{}) bool { x, f := v.(Register32)       ; return f && x != WSP }
func isXrOrSP    (v interface{}) bool { x, f := v.(Register64)       ; return f && x != XZR }
func isWrOrWSP   (v interface{}) bool { x, f := v.(Register32)       ; return f && x != WZR }
func isBr        (v interface{}) bool { _, f := v.(SIMDRegister8)    ; return f }
func isHr        (v interface{}) bool { _, f := v.(SIMDRegister16)   ; return f }
func isSr        (v interface{}) bool { _, f := v.(SIMDRegister32)   ; return f }
func isDr        (v interface{}) bool { _, f := v.(SIMDRegister64)   ; return f }
func isQr        (v interface{}) bool { _, f := v.(SIMDRegister128)  ; return f }
func isVr        (v interface{}) bool { _, f := v.(SIMDRegister128v) ; return f }
func isVri       (v interface{}) bool { _, f := v.(_Indexed128r)     ; return f }

func isImm       (v interface{}) bool { _, f := asInt64(v)  ; return f }
func isImm9      (v interface{}) bool { x, f := asInt64(v)  ; return f && x &^ 0b111111111 == 0 }
func isImm12     (v interface{}) bool { x, f := asInt64(v)  ; return f && x &^ 0b111111111111 == 0 }
func isUimm4     (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0b1111 == 0 }
func isUimm5     (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0b11111 == 0 }
func isUimm6     (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0b111111 == 0 }
func isUimm7     (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0b1111111 == 0 }
func isUimm8     (v interface{}) bool { x, f := asUint64(v) ; return f && x <= math.MaxUint8 }
func isUimm16    (v interface{}) bool { x, f := asUint64(v) ; return f && x <= math.MaxUint16 }
func isMask32    (v interface{}) bool { x, f := asUint64(v) ; return f && _BitMask(x).is32() }
func isMask64    (v interface{}) bool { x, f := asUint64(v) ; return f && _BitMask(x).is64() }
func isFpImm8    (v interface{}) bool { _, f := asFloat8(v) ; return f }
func isFpBits    (v interface{}) bool { x, f := asUint64(v) ; return f && x >= 1 && x <= 64 }

func isMod       (v interface{}) bool { _, f := v.(Modifier)        ; return f }
func isIndex     (v interface{}) bool { _, f := v.(IndexMode)       ; return f }
func isShift     (v interface{}) bool { _, f := v.(ShiftType)       ; return f }
func isPState    (v interface{}) bool { _, f := v.(PStateField)     ; return f }
func isBrCond    (v interface{}) bool { _, f := v.(BranchCondition) ; return f }
func isSysReg    (v interface{}) bool { _, f := v.(SystemRegister)  ; return f }
func isExtend    (v interface{}) bool { _, f := v.(Extension)       ; return f }
func isOption    (v interface{}) bool { _, f := v.(BarrierOption)   ; return f }
func isOptionNXS (v interface{}) bool { x, f := v.(BarrierOption)   ; return f && x.isNXS() }
func isTargets   (v interface{}) bool { _, f := v.(BranchTarget)    ; return f }
func isRangePrf  (v interface{}) bool { _, f := v.(RangePrefetchOp) ; return f }

func isMem(v interface{}) bool {
    if x, ok := v.(*asm.MemoryOperand); !ok {
        return false
    } else {
        _, ok = x.Addr.(asm.MemoryAddress)
        return ok
    }
}

func isVfmt(v interface{}, fmt ...SIMDVectorArrangement) bool {
    t := vfmt(v)
    for _, f := range fmt { if t == f { return true } }
    return false
}

func isIntLit(v interface{}, imm ...int64) bool {
    var t int64
    var x rt.GoEface

    /* check for special values */
    if isSpecial(v) {
        return false
    }

    /* check for integer types */
    if x = rt.AsEface(v); !isInt(x.Kind()) && !isUint(x.Kind()) {
        return false
    }

    /* find the literal */
    t = x.ToInt64()
    for _, n := range imm { if t == n { return true } }
    return false
}

func isFloatLit(v interface{}, imm float64) bool {
    if isSpecial(v) {
        return false
    } else if x := rt.AsEface(v); !isFloat(x.Kind()) {
        return false
    } else {
        return x.ToFloat64() == imm
    }
}

func isSameMod(v interface{}, mod Modifier) bool {
    if _, ok := mod.(LSL); ok && v == LSL12 {
        return true
    } else {
        return rt.TypeOf(v) == rt.TypeOf(mod)
    }
}

func isWrOrXr(v interface{}) bool {
    switch v.(type) {
        case Register32 : return true
        case Register64 : return true
        default         : return false
    }
}

func isAdvSIMD(v interface{}) bool {
    switch v.(type) {
        case SIMDRegister8   : return true
        case SIMDRegister16  : return true
        case SIMDRegister32  : return true
        case SIMDRegister64  : return true
        case SIMDRegister128 : return true
        default              : return false
    }
}

func isNextReg(rd, rs interface{}, dist uint8) bool {
    return isSameType(rs, rd) && rd.(asm.Register).ID() == (rs.(asm.Register).ID() + dist) % 32
}
