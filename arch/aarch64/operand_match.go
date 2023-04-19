package aarch64

import (
    `math`
    `reflect`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/rt`
)

const (
    _MaxInt12 = 2047
    _MinInt12 = -2048
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

var _TLBIPOptions = map[TLBIOption]bool {
    VAE1OS         : true,
    VAAE1OS        : true,
    VALE1OS        : true,
    VAALE1OS       : true,
    RVAE1IS        : true,
    RVAAE1IS       : true,
    RVALE1IS       : true,
    RVAALE1IS      : true,
    VAE1IS         : true,
    VAAE1IS        : true,
    VALE1IS        : true,
    VAALE1IS       : true,
    RVAE1OS        : true,
    RVAAE1OS       : true,
    RVALE1OS       : true,
    RVAALE1OS      : true,
    RVAE1          : true,
    RVAAE1         : true,
    RVALE1         : true,
    RVAALE1        : true,
    VAE1           : true,
    VAAE1          : true,
    VALE1          : true,
    VAALE1         : true,
    VAE1OSNXS      : true,
    VAAE1OSNXS     : true,
    VALE1OSNXS     : true,
    VAALE1OSNXS    : true,
    RVAE1ISNXS     : true,
    RVAAE1ISNXS    : true,
    RVALE1ISNXS    : true,
    RVAALE1ISNXS   : true,
    VAE1ISNXS      : true,
    VAAE1ISNXS     : true,
    VALE1ISNXS     : true,
    VAALE1ISNXS    : true,
    RVAE1OSNXS     : true,
    RVAAE1OSNXS    : true,
    RVALE1OSNXS    : true,
    RVAALE1OSNXS   : true,
    RVAE1NXS       : true,
    RVAAE1NXS      : true,
    RVALE1NXS      : true,
    RVAALE1NXS     : true,
    VAE1NXS        : true,
    VAAE1NXS       : true,
    VALE1NXS       : true,
    VAALE1NXS      : true,
    IPAS2E1IS      : true,
    RIPAS2E1IS     : true,
    IPAS2LE1IS     : true,
    RIPAS2LE1IS    : true,
    VAE2OS         : true,
    VALE2OS        : true,
    RVAE2IS        : true,
    RVALE2IS       : true,
    VAE2IS         : true,
    VALE2IS        : true,
    IPAS2E1OS      : true,
    IPAS2E1        : true,
    RIPAS2E1       : true,
    RIPAS2E1OS     : true,
    IPAS2LE1OS     : true,
    IPAS2LE1       : true,
    RIPAS2LE1      : true,
    RIPAS2LE1OS    : true,
    RVAE2OS        : true,
    RVALE2OS       : true,
    RVAE2          : true,
    RVALE2         : true,
    VAE2           : true,
    VALE2          : true,
    IPAS2E1ISNXS   : true,
    RIPAS2E1ISNXS  : true,
    IPAS2LE1ISNXS  : true,
    RIPAS2LE1ISNXS : true,
    VAE2OSNXS      : true,
    VALE2OSNXS     : true,
    RVAE2ISNXS     : true,
    RVALE2ISNXS    : true,
    VAE2ISNXS      : true,
    VALE2ISNXS     : true,
    IPAS2E1OSNXS   : true,
    IPAS2E1NXS     : true,
    RIPAS2E1NXS    : true,
    RIPAS2E1OSNXS  : true,
    IPAS2LE1OSNXS  : true,
    IPAS2LE1NXS    : true,
    RIPAS2LE1NXS   : true,
    RIPAS2LE1OSNXS : true,
    RVAE2OSNXS     : true,
    RVALE2OSNXS    : true,
    RVAE2NXS       : true,
    RVALE2NXS      : true,
    VAE2NXS        : true,
    VALE2NXS       : true,
    VAE3OS         : true,
    VALE3OS        : true,
    RVAE3IS        : true,
    RVALE3IS       : true,
    VAE3IS         : true,
    VALE3IS        : true,
    RVAE3OS        : true,
    RVALE3OS       : true,
    RVAE3          : true,
    RVALE3         : true,
    VAE3           : true,
    VALE3          : true,
    VAE3OSNXS      : true,
    VALE3OSNXS     : true,
    RVAE3ISNXS     : true,
    RVALE3ISNXS    : true,
    VAE3ISNXS      : true,
    VALE3ISNXS     : true,
    RVAE3OSNXS     : true,
    RVALE3OSNXS    : true,
    RVAE3NXS       : true,
    RVALE3NXS      : true,
    VAE3NXS        : true,
    VALE3NXS       : true,
}

func isInt(k reflect.Kind) bool {
    return (_IntMask & (1 << k)) != 0
}

func isUint(k reflect.Kind) bool {
    return (_UintMask & (1 << k)) != 0
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
        case WRegister             : return true
        case XRegister             : return true
        case BRegister             : return true
        case HRegister             : return true
        case SRegister             : return true
        case DRegister             : return true
        case QRegister             : return true
        case VRegister             : return true
        case VidxRegister          : return true
        case SystemRegister        : return true
        case VecFormat             : return true
        case Vector                : return true
        case IndexedVector         : return true
        case IndexMode             : return true
        case MSL                   : return true
        case LSL                   : return true
        case LSR                   : return true
        case ASR                   : return true
        case ROR                   : return true
        case UXTB                  : return true
        case UXTH                  : return true
        case UXTW                  : return true
        case UXTX                  : return true
        case SXTB                  : return true
        case SXTH                  : return true
        case SXTW                  : return true
        case SXTX                  : return true
        case ModType               : return true
        case BranchTarget          : return true
        case ConditionCode         : return true
        case BarrierOption         : return true
        case PStateField           : return true
        case ATOption              : return true
        case BRBOption             : return true
        case DCOption              : return true
        case ICOption              : return true
        case TLBIOption            : return true
        case SMEOption             : return true
        case Symbol                : return true
        case PrefetchOp            : return true
        case RangePrefetchOp       : return true
        case asm.Feature           : return true
        case asm.BranchType        : return true
        case asm.PseudoType        : return true
        case asm.RelativeOffset    : return true
        case asm.InstructionDomain : return true
        default                    : return false
    }
}

func isSameType(x, y interface{}) bool {
    return rt.TypeOf(x) == rt.TypeOf(y)
}

func isLabel        (v interface{}) bool { _, f := v.(*asm.Label)    ; return f }
func isXr           (v interface{}) bool { x, f := v.(XRegister)     ; return f && x != SP }
func isWr           (v interface{}) bool { x, f := v.(WRegister)     ; return f && x != WSP }
func isXrOrSP       (v interface{}) bool { x, f := v.(XRegister)     ; return f && x != XZR }
func isWrOrWSP      (v interface{}) bool { x, f := v.(WRegister)     ; return f && x != WZR }
func isBr           (v interface{}) bool { _, f := v.(BRegister)     ; return f }
func isHr           (v interface{}) bool { _, f := v.(HRegister)     ; return f }
func isSr           (v interface{}) bool { _, f := v.(SRegister)     ; return f }
func isDr           (v interface{}) bool { _, f := v.(DRegister)     ; return f }
func isQr           (v interface{}) bool { _, f := v.(QRegister)     ; return f }
func isVr           (v interface{}) bool { _, f := v.(VRegister)     ; return f }
func isVri          (v interface{}) bool { _, f := v.(VidxRegister)  ; return f }
func isVec1         (v interface{}) bool { x, f := v.(Vector)        ; return f && x.Size() == 1 }
func isVec2         (v interface{}) bool { x, f := v.(Vector)        ; return f && x.Size() == 2 }
func isVec3         (v interface{}) bool { x, f := v.(Vector)        ; return f && x.Size() == 3 }
func isVec4         (v interface{}) bool { x, f := v.(Vector)        ; return f && x.Size() == 4 }
func isIdxVec1      (v interface{}) bool { x, f := v.(IndexedVector) ; return f && x.Size() == 1 }
func isIdxVec2      (v interface{}) bool { x, f := v.(IndexedVector) ; return f && x.Size() == 2 }
func isIdxVec3      (v interface{}) bool { x, f := v.(IndexedVector) ; return f && x.Size() == 3 }
func isIdxVec4      (v interface{}) bool { x, f := v.(IndexedVector) ; return f && x.Size() == 4 }

func isImm8         (v interface{}) bool { x, f := asInt64(v)  ; return f && x >= math.MinInt8 && x <= math.MaxInt8 }
func isImm12        (v interface{}) bool { x, f := asInt64(v)  ; return f && x >= _MinInt12 && x <= _MaxInt12 }
func isUimm3        (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0x07 == 0 }
func isUimm4        (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0x0f == 0 }
func isUimm5        (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0x1f == 0 }
func isUimm6        (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0x3f == 0 }
func isUimm7        (v interface{}) bool { x, f := asUint64(v) ; return f && x &^ 0x7f == 0 }
func isUimm8        (v interface{}) bool { x, f := asUint64(v) ; return f && x <= math.MaxUint8 }
func isUimm16       (v interface{}) bool { x, f := asUint64(v) ; return f && x <= math.MaxUint16 }
func isMask32       (v interface{}) bool { x, f := asUint64(v) ; return f && _BitMask(x).is32() }
func isMask64       (v interface{}) bool { x, f := asUint64(v) ; return f && _BitMask(x).is64() }
func isFpBits       (v interface{}) bool { x, f := asUint64(v) ; return f && x >= 1 && x <= 64 }

func isMod          (v interface{}) bool { _, f := v.(Modifier)        ; return f }
func isIndex        (v interface{}) bool { _, f := v.(IndexMode)       ; return f }
func isPState       (v interface{}) bool { _, f := v.(PStateField)     ; return f }
func isBrCond       (v interface{}) bool { _, f := v.(ConditionCode)   ; return f }
func isSysReg       (v interface{}) bool { _, f := v.(SystemRegister)  ; return f }
func isOption       (v interface{}) bool { _, f := v.(BarrierOption)   ; return f || isUimm4(v) }
func isOptionNXS    (v interface{}) bool { x, f := v.(BarrierOption)   ; return f && x.isNXS() }
func isTargets      (v interface{}) bool { _, f := v.(BranchTarget)    ; return f }
func isBasicPrf     (v interface{}) bool { _, f := v.(PrefetchOp)      ; return f || isUimm5(v) }
func isRangePrf     (v interface{}) bool { _, f := v.(RangePrefetchOp) ; return f || isUimm6(v) }
func isATOption     (v interface{}) bool { _, f := v.(ATOption)        ; return f }
func isBRBOption    (v interface{}) bool { _, f := v.(BRBOption)       ; return f }
func isDCOption     (v interface{}) bool { _, f := v.(DCOption)        ; return f }
func isICOption     (v interface{}) bool { _, f := v.(ICOption)        ; return f }
func isSMEOption    (v interface{}) bool { _, f := v.(SMEOption)       ; return f }
func isTLBIOption   (v interface{}) bool { _, f := v.(TLBIOption)      ; return f }
func isTLBIPOption  (v interface{}) bool { x, f := v.(TLBIOption)      ; return f && _TLBIPOptions[x] }

func isMem(v interface{}) bool {
    if x, ok := v.(*asm.MemoryOperand); !ok {
        return false
    } else {
        _, ok = x.Addr.(asm.MemoryAddress)
        return ok
    }
}

func isMods(v interface{}, mod ...ModType) bool {
    t := modt(v)
    for _, f := range mod { if t == f { return true } }
    return false
}

func isVfmt(v interface{}, fmt ...VecFormat) bool {
    t := vfmt(v)
    for _, f := range fmt { if t == f { return true } }
    return false
}

func isFpImm8(v interface{}) bool {
    _, f := encodeFpImm8(v)
    return f
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

func isExtIndex(r, v interface{}) bool {
    return isUimm4(v) && (vfmt(r) != Vec8B || (asUimm4(v) & 0b1000) == 0)
}

func isFloatLit(v interface{}, imm float64) bool {
    x, ok := asFloat64(v)
    return ok && x == imm
}

func isWrOrXr(v interface{}) bool {
    switch x := v.(type) {
        case XRegister : return x != SP
        case WRegister : return x != WSP
        default        : return false
    }
}

func isAdvSIMD(v interface{}) bool {
    switch v.(type) {
        case BRegister : return true
        case HRegister : return true
        case SRegister : return true
        case DRegister : return true
        case QRegister : return true
        default        : return false
    }
}

func isNextReg(rd, rs interface{}, dist uint8) bool {
    return isSameType(rs, rd) && rd.(asm.Register).ID() == (rs.(asm.Register).ID() + dist) % 32
}

func isMOVxImm(v interface{}, bits uint64, inverted bool) bool {
    _, ok := encodeMoveImm(v, bits, inverted)
    return ok
}

func isBFxWidth(i, w interface{}, n uint64) bool {
    if vi, ok := asUint64(i); !ok {
        return false
    } else if vw, ok := asUint64(w); !ok {
        return false
    } else {
        return vw >= 1 && vi <= n - 1 && vw <= n - vi
    }
}
