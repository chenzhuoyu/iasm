package aarch64

import (
    `fmt`
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

type (
    _MemOpExt struct{}
    _Basic    struct{}
)

// MemOpExt is the aarch specific extensions of asm.MemoryOperand
var MemOpExt _MemOpExt

func (_MemOpExt) Free()                   {}
func (_MemOpExt) Sealed(tag.Tag)          {}
func (_MemOpExt) MemoryOperandExtension() {}

func (_MemOpExt) String(mem *asm.MemoryOperand) string {
    return mem.Addr.String()
}

func (_MemOpExt) EnsureValid(mem *asm.MemoryOperand) {
    mem.Addr.EnsureValid()
}

// Basic reprensets a basic memory address that does not have any indexing or extensions.
var Basic _Basic

func (_Basic) Free()                   {}
func (_Basic) Sealed(_ tag.Tag)        {}
func (_Basic) MemoryAddressExtension() {}

func (_Basic) String(addr asm.MemoryAddress) string {
    sb := new(strings.Builder)
    sb.WriteByte('[')
    sb.WriteString(addr.Base.String())

    /* register offset */
    if addr.Index != nil {
        sb.WriteString(", ")
        sb.WriteString(addr.Index.String())
    }

    /* immediate offset */
    if addr.Offset != 0 {
        sb.WriteString(", ")
        sb.WriteString(fmt.Sprintf("#%d", addr.Offset))
    }

    /* compose the address */
    sb.WriteByte(']')
    return sb.String()
}

func (_Basic) EnsureValid(addr asm.MemoryAddress) {
    if addr.Index != nil && addr.Offset != 0 {
        panic("aarch64: cannot index base by register and immediate at the same time")
    }
}

// IndexMode represents the memory operand index mode.
type IndexMode uint8

const (
    PreIndex IndexMode = iota   // Pre-index, like "[<Xn|SP>, #imm]!"
    PostIndex                   // Post-index, like "[<Xn|SP>], #imm"
)

func (IndexMode) Free()                   {}
func (IndexMode) Sealed(_ tag.Tag)        {}
func (IndexMode) MemoryAddressExtension() {}

func (self IndexMode) String(addr asm.MemoryAddress) string {
    sb := new(strings.Builder)
    sb.WriteByte('[')
    sb.WriteString(addr.Base.String())

    /* register offset */
    if addr.Index != nil {
        sb.WriteString(", ")
        sb.WriteString(addr.Index.String())
    }

    /* the offset of post-index does not appear here */
    if addr.Offset != 0 && self != PostIndex {
        sb.WriteString(", ")
        sb.WriteString(fmt.Sprintf("#%d", addr.Offset))
    }

    /* check for index type */
    switch self {
        case PreIndex  : sb.WriteString("]!")
        case PostIndex : sb.WriteString("], ")
        default        : panic("unreachable")
    }

    /* not post-index, no more text */
    if self != PostIndex {
        return sb.String()
    }

    /* format the post-index offset */
    sb.WriteString(fmt.Sprintf("#%d", addr.Offset))
    return sb.String()
}

func (self IndexMode) EnsureValid(addr asm.MemoryAddress) {
    if addr.Index != nil && addr.Offset != 0 {
        panic("aarch64: cannot index base by register and immediate at the same time")
    }
}

// Modifier represents one of the memory address modifiers, such as shifts and extends.
type Modifier interface {
    asm.MemoryAddressExtension
    Name() string
    Amount() uint8
}

func _Modifier_String(mod Modifier, addr asm.MemoryAddress) string {
    sb := new(strings.Builder)
    sb.WriteByte('[')
    sb.WriteString(addr.Base.String())

    /* register offset */
    if addr.Index != nil {
        sb.WriteString(", ")
        sb.WriteString(addr.Index.String())
    }

    /* immediate offset */
    if addr.Offset != 0 {
        sb.WriteString(", ")
        sb.WriteString(fmt.Sprintf("#%d", addr.Offset))
    }

    /* construct the memory address */
    sb.WriteString(mod.Name())
    sb.WriteString(fmt.Sprintf(" %d]", mod.Amount()))
    return sb.String()
}

func _Modifier_EnsureValid(addr asm.MemoryAddress) {
    if addr.Index != nil && addr.Offset != 0 {
        panic("aarch64: cannot index base by register and immediate at the same time")
    }
}

// ShiftType represents one of the register shifts.
type ShiftType interface {
    Modifier
    ShiftType() uint8
}

type (
    MSL uint8   // Masking shift left
    LSL uint8   // Logical shift left
    LSR uint8   // Logical shift right
    ASR uint8   // Arithmetic shift right
    ROR uint8   // Rotate right
)

func (MSL) Free() {}
func (LSL) Free() {}
func (LSR) Free() {}
func (ASR) Free() {}
func (ROR) Free() {}

func (MSL) Sealed(_ tag.Tag) {}
func (LSL) Sealed(_ tag.Tag) {}
func (LSR) Sealed(_ tag.Tag) {}
func (ASR) Sealed(_ tag.Tag) {}
func (ROR) Sealed(_ tag.Tag) {}

func (MSL) MemoryAddressExtension() {}
func (LSL) MemoryAddressExtension() {}
func (LSR) MemoryAddressExtension() {}
func (ASR) MemoryAddressExtension() {}
func (ROR) MemoryAddressExtension() {}

func (MSL) Name() string { return "MSL" }
func (LSL) Name() string { return "LSL" }
func (LSR) Name() string { return "LSR" }
func (ASR) Name() string { return "ASR" }
func (ROR) Name() string { return "ROR" }

func (self MSL) Amount() uint8 { return uint8(self) }
func (self LSL) Amount() uint8 { return uint8(self) }
func (self LSR) Amount() uint8 { return uint8(self) }
func (self ASR) Amount() uint8 { return uint8(self) }
func (self ROR) Amount() uint8 { return uint8(self) }

func (MSL) ShiftType() uint8 { return 0b00 }
func (LSL) ShiftType() uint8 { return 0b00 }
func (LSR) ShiftType() uint8 { return 0b01 }
func (ASR) ShiftType() uint8 { return 0b10 }
func (ROR) ShiftType() uint8 { return 0b11 }

func (self MSL) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self LSL) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self LSR) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self ASR) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self ROR) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }

func (MSL) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (LSL) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (LSR) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (ASR) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (ROR) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }

type (
    _LSL12 struct{}
)

// LSL12 shifts the immediate value by 12, this value can only be used with {ADD,SUB}[S] instructions.
var LSL12 _LSL12

func (_LSL12) Free() {}
func (_LSL12) Sealed(_ tag.Tag) {}
func (_LSL12) MemoryAddressExtension()              {}
func (_LSL12) Name() string                         { return "LSL #12" }
func (_LSL12) Amount() uint8                        { return 0 }
func (_LSL12) String(addr asm.MemoryAddress) string { return _Modifier_String(LSL12, addr) }
func (_LSL12) ShiftType() uint8                     { return 0b1 }
func (_LSL12) EnsureValid(addr asm.MemoryAddress)   { _Modifier_EnsureValid(addr) }

// Extension represents one of the register extensions.
type Extension interface {
    Modifier
    Extension() uint8
}

type (
    UXTB uint8  // Unsigned extension from byte to 64-bit
    UXTH uint8  // Unsigned extension from half-word to 64-bit
    UXTW uint8  // Unsigned extension from word to 64-bit
    UXTX uint8  // Unsigned extension from 64-bit to 64-bit
    SXTB uint8  // Signed extension from byte to 64-bit
    SXTH uint8  // Signed extension from half-word to 64-bit
    SXTW uint8  // Signed extension from word to 64-bit
    SXTX uint8  // Signed extension from 64-bit to 64-bit
)

func (UXTB) Free() {}
func (UXTH) Free() {}
func (UXTW) Free() {}
func (UXTX) Free() {}
func (SXTB) Free() {}
func (SXTH) Free() {}
func (SXTW) Free() {}
func (SXTX) Free() {}

func (UXTB) Sealed(_ tag.Tag) {}
func (UXTH) Sealed(_ tag.Tag) {}
func (UXTW) Sealed(_ tag.Tag) {}
func (UXTX) Sealed(_ tag.Tag) {}
func (SXTB) Sealed(_ tag.Tag) {}
func (SXTH) Sealed(_ tag.Tag) {}
func (SXTW) Sealed(_ tag.Tag) {}
func (SXTX) Sealed(_ tag.Tag) {}

func (UXTB) MemoryAddressExtension() {}
func (UXTH) MemoryAddressExtension() {}
func (UXTW) MemoryAddressExtension() {}
func (UXTX) MemoryAddressExtension() {}
func (SXTB) MemoryAddressExtension() {}
func (SXTH) MemoryAddressExtension() {}
func (SXTW) MemoryAddressExtension() {}
func (SXTX) MemoryAddressExtension() {}

func (UXTB) Name() string { return "UXTB" }
func (UXTH) Name() string { return "UXTH" }
func (UXTW) Name() string { return "UXTW" }
func (UXTX) Name() string { return "UXTX" }
func (SXTB) Name() string { return "SXTB" }
func (SXTH) Name() string { return "SXTH" }
func (SXTW) Name() string { return "SXTW" }
func (SXTX) Name() string { return "SXTX" }

func (self UXTB) Amount() uint8 { return uint8(self) }
func (self UXTH) Amount() uint8 { return uint8(self) }
func (self UXTW) Amount() uint8 { return uint8(self) }
func (self UXTX) Amount() uint8 { return uint8(self) }
func (self SXTB) Amount() uint8 { return uint8(self) }
func (self SXTH) Amount() uint8 { return uint8(self) }
func (self SXTW) Amount() uint8 { return uint8(self) }
func (self SXTX) Amount() uint8 { return uint8(self) }

func (LSL)  Extension() uint8 { return 0xff }
func (UXTB) Extension() uint8 { return 0b000 }
func (UXTH) Extension() uint8 { return 0b001 }
func (UXTW) Extension() uint8 { return 0b010 }
func (UXTX) Extension() uint8 { return 0b011 }
func (SXTB) Extension() uint8 { return 0b100 }
func (SXTH) Extension() uint8 { return 0b101 }
func (SXTW) Extension() uint8 { return 0b110 }
func (SXTX) Extension() uint8 { return 0b111 }

func (self UXTB) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self UXTH) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self UXTW) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self UXTX) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTB) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTH) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTW) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTX) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }

func (UXTB) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (UXTH) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (UXTW) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (UXTX) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTB) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTH) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTW) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTX) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }

// Mem constructs a memory operand.
func Mem(base asm.Register, args ...interface{}) (v *asm.MemoryOperand) {
    addr := asm.MemoryAddress{}
    addr.Ext = Basic
    addr.Base = base

    /* parse offset or index register if any */
    if len(args) >= 1 {
        if isInt32(args[0]) {
            addr.Offset = asInt32(args[0])
        } else if rr, ok := args[0].(asm.Register); ok {
            addr.Index = rr
        } else {
            panic("aarch64: invalid memory offset or index")
        }
    }

    /* parse extensions if any */
    if len(args) >= 2 {
        if isMod(args[1]) || isIndex(args[1]) {
            addr.Ext = args[1].(asm.MemoryAddressExtension)
        } else {
            panic("aarch64: invalid memory index, shift or extension")
        }
    }

    /* more arguments than required */
    if len(args) >= 3 {
        panic("aarch64: excessive arguments for memory operand")
    }

    /* construct the operand */
    v = asm.CreateMemoryOperand(MemOpExt)
    v.Addr = addr
    v.EnsureValid()
    return
}

// BranchTarget represents the operand for BTI instruction.
type BranchTarget uint8

const (
    _BrOmitted BranchTarget = iota
    BrC
    BrJ
    BrJC
)

// BranchCondition represents one of the conditional branch conditions.
type BranchCondition uint8

const (
    EQ BranchCondition = 0b0000     // Equal / equals zero                      Z != 0
    NE BranchCondition = 0b0001     // Not equal                                Z == 0
    CS BranchCondition = 0b0010     // Carry set / unsigned higher or same      C != 0
    CC BranchCondition = 0b0011     // Carry clear / unsigned lower             C == 0
    MI BranchCondition = 0b0100     // Minus / negative                         N != 0
    PL BranchCondition = 0b0101     // Plus / positive or zero                  N == 0
    VS BranchCondition = 0b0110     // Overflow                                 V != 0
    VC BranchCondition = 0b0111     // No overflow                              V == 0
    HI BranchCondition = 0b1000     // Unsigned higher                          C != 0 and Z == 0
    LS BranchCondition = 0b1001     // Unsigned lower or same                   C == 0 or Z != 0
    GE BranchCondition = 0b1010     // Signed greater than or equal             N == V
    LT BranchCondition = 0b1011     // Signed less than                         N != V
    GT BranchCondition = 0b1100     // Signed greater than                      Z == 0 and N == V
    LE BranchCondition = 0b1101     // Signed less than or equal                Z != 0 or N != V
    AL BranchCondition = 0b1110     // Always (default)                         any
)

const (
    HS = CS                         // Unsigned higher or same / carry set      C != 0
    LO = CC                         // Unsigned lower / carry clear             C == 0
)

// BarrierOption represents one of the memory barrier options.
type BarrierOption uint8

const (
    // SY : Full system is the required shareability domain, reads and writes are the required
    // access types, both before and after the barrier instruction. This option is referred to
    // as the full system barrier.
    SY BarrierOption = 0b1111

    // ST : Full system is the required shareability domain, writes are the required access type,
    // both before and after the barrier instruction.
    ST BarrierOption = 0b1110

    // LD : Full system is the required shareability domain, reads are the required access type
    // before the barrier instruction, and reads and writes are the required access types after
    // the barrier instruction.
    LD BarrierOption = 0b1101

    // ISH : Inner Shareable is the required shareability domain, reads and writes are the required
    // access types, both before and after the barrier instruction.
    ISH BarrierOption = 0b1011

    // ISHST : Inner Shareable is the required shareability domain, writes are the required access
    // type, both before and after the barrier instruction.
    ISHST BarrierOption = 0b1010

    // ISHLD : Inner Shareable is the required shareability domain, reads are the required access
    // type before the barrier instruction, and reads and writes are the required access types after
    // the barrier instruction.
    ISHLD BarrierOption = 0b1001

    // NSH : Non-shareable is the required shareability domain, reads and writes are the required
    // access, both before and after the barrier instruction.
    NSH BarrierOption = 0b0111

    // NSHST : Non-shareable is the required shareability domain, writes are the required access
    // type, both before and after the barrier instruction.
    NSHST BarrierOption = 0b0110

    // NSHLD : Non-shareable is the required shareability domain, reads are the required access
    // type before the barrier instruction, and reads and writes are the required access types after
    // the barrier instruction.
    NSHLD BarrierOption = 0b0101

    // OSH : Outer Shareable is the required shareability domain, reads and writes are the required
    // access types, both before and after the barrier instruction.
    OSH BarrierOption = 0b0011

    // OSHST : Outer Shareable is the required shareability domain, writes are the required access
    // type, both before and after the barrier instruction.
    OSHST BarrierOption = 0b0010

    // OSHLD : Outer Shareable is the required shareability domain, reads are the required access
    // type before the barrier instruction, and reads and writes are the required access types after
    // the barrier instruction.
    OSHLD BarrierOption = 0b0001
)
