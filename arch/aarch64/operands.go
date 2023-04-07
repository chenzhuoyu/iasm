package aarch64

import (
    `fmt`
    `strings`

    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

type (
    _Basic    struct{}
    _MemOpExt struct{}
)

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

// ModType represents one of the modifier types.
type ModType uint8

const (
    ModInvalid ModType = iota
    ModMSL
    ModLSL
    ModLSR
    ModASR
    ModROR
    ModUXTB
    ModUXTH
    ModUXTW
    ModUXTX
    ModSXTB
    ModSXTH
    ModSXTW
    ModSXTX
)

func (self ModType) String() string {
    switch self {
        case ModMSL  : return "MSL"
        case ModLSL  : return "LSL"
        case ModLSR  : return "LSR"
        case ModASR  : return "ASR"
        case ModROR  : return "ROR"
        case ModUXTB : return "UXTB"
        case ModUXTH : return "UXTH"
        case ModUXTW : return "UXTW"
        case ModUXTX : return "UXTX"
        case ModSXTB : return "SXTB"
        case ModSXTH : return "SXTH"
        case ModSXTW : return "SXTW"
        case ModSXTX : return "SXTX"
        default      : return "???"
    }
}

// Modifier represents one of the memory address modifiers, such as shifts and extends.
type Modifier interface {
    asm.MemoryAddressExtension
    Type() ModType
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
    sb.WriteString(mod.Type().String())
    sb.WriteString(fmt.Sprintf(" %d]", mod.Amount()))
    return sb.String()
}

func _Modifier_EnsureValid(addr asm.MemoryAddress) {
    if addr.Index != nil && addr.Offset != 0 {
        panic("aarch64: cannot index base by register and immediate at the same time")
    }
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

func (MSL) Type() ModType { return ModMSL }
func (LSL) Type() ModType { return ModLSL }
func (LSR) Type() ModType { return ModLSR }
func (ASR) Type() ModType { return ModASR }
func (ROR) Type() ModType { return ModROR }

func (MSL) Sealed(_ tag.Tag) {}
func (LSL) Sealed(_ tag.Tag) {}
func (LSR) Sealed(_ tag.Tag) {}
func (ASR) Sealed(_ tag.Tag) {}
func (ROR) Sealed(_ tag.Tag) {}

func (MSL) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (LSL) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (LSR) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (ASR) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (ROR) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }

func (MSL) MemoryAddressExtension() {}
func (LSL) MemoryAddressExtension() {}
func (LSR) MemoryAddressExtension() {}
func (ASR) MemoryAddressExtension() {}
func (ROR) MemoryAddressExtension() {}

func (self MSL) Amount() uint8 { return uint8(self) }
func (self LSL) Amount() uint8 { return uint8(self) }
func (self LSR) Amount() uint8 { return uint8(self) }
func (self ASR) Amount() uint8 { return uint8(self) }
func (self ROR) Amount() uint8 { return uint8(self) }

func (self MSL) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self LSL) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self LSR) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self ASR) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self ROR) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }

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

func (UXTB) Type() ModType { return ModUXTB }
func (UXTH) Type() ModType { return ModUXTH }
func (UXTW) Type() ModType { return ModUXTW }
func (UXTX) Type() ModType { return ModUXTX }
func (SXTB) Type() ModType { return ModSXTB }
func (SXTH) Type() ModType { return ModSXTH }
func (SXTW) Type() ModType { return ModSXTW }
func (SXTX) Type() ModType { return ModSXTX }

func (UXTB) Sealed(_ tag.Tag) {}
func (UXTH) Sealed(_ tag.Tag) {}
func (UXTW) Sealed(_ tag.Tag) {}
func (UXTX) Sealed(_ tag.Tag) {}
func (SXTB) Sealed(_ tag.Tag) {}
func (SXTH) Sealed(_ tag.Tag) {}
func (SXTW) Sealed(_ tag.Tag) {}
func (SXTX) Sealed(_ tag.Tag) {}

func (UXTB) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (UXTH) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (UXTW) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (UXTX) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTB) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTH) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTW) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }
func (SXTX) EnsureValid(addr asm.MemoryAddress) { _Modifier_EnsureValid(addr) }

func (UXTB) MemoryAddressExtension() {}
func (UXTH) MemoryAddressExtension() {}
func (UXTW) MemoryAddressExtension() {}
func (UXTX) MemoryAddressExtension() {}
func (SXTB) MemoryAddressExtension() {}
func (SXTH) MemoryAddressExtension() {}
func (SXTW) MemoryAddressExtension() {}
func (SXTX) MemoryAddressExtension() {}

func (self UXTB) Amount() uint8 { return uint8(self) }
func (self UXTH) Amount() uint8 { return uint8(self) }
func (self UXTW) Amount() uint8 { return uint8(self) }
func (self UXTX) Amount() uint8 { return uint8(self) }
func (self SXTB) Amount() uint8 { return uint8(self) }
func (self SXTH) Amount() uint8 { return uint8(self) }
func (self SXTW) Amount() uint8 { return uint8(self) }
func (self SXTX) Amount() uint8 { return uint8(self) }

func (self UXTB) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self UXTH) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self UXTW) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self UXTX) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTB) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTH) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTW) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }
func (self SXTX) String(addr asm.MemoryAddress) string { return _Modifier_String(self, addr) }

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

// ConditionCode represents one of the condition codes.
type ConditionCode uint8

const (
    EQ ConditionCode = 0b0000       // Equal / equals zero                      Z != 0
    NE ConditionCode = 0b0001       // Not equal                                Z == 0
    CS ConditionCode = 0b0010       // Carry set / unsigned higher or same      C != 0
    CC ConditionCode = 0b0011       // Carry clear / unsigned lower             C == 0
    MI ConditionCode = 0b0100       // Minus / negative                         N != 0
    PL ConditionCode = 0b0101       // Plus / positive or zero                  N == 0
    VS ConditionCode = 0b0110       // Overflow                                 V != 0
    VC ConditionCode = 0b0111       // No overflow                              V == 0
    HI ConditionCode = 0b1000       // Unsigned higher                          C != 0 and Z == 0
    LS ConditionCode = 0b1001       // Unsigned lower or same                   C == 0 or Z != 0
    GE ConditionCode = 0b1010       // Signed greater than or equal             N == V
    LT ConditionCode = 0b1011       // Signed less than                         N != V
    GT ConditionCode = 0b1100       // Signed greater than                      Z == 0 and N == V
    LE ConditionCode = 0b1101       // Signed less than or equal                Z != 0 or N != V
    AL ConditionCode = 0b1110       // Always (default)                         any
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

func (self BarrierOption) nxs() uint32 {
    switch self {
        case SY  : return 0b11
        case ISH : return 0b10
        case NSH : return 0b01
        case OSH : return 0b00
        default  : panic("aarch64: invalid nXS barrier option")
    }
}

func (self BarrierOption) isNXS() bool {
    switch self {
        case SY  : return true
        case ISH : return true
        case NSH : return true
        case OSH : return true
        default  : return false
    }
}

// PStateField represents one of PSTATE field combinations.
type PStateField uint8

const (
    UAO      PStateField = 0b000_011
    PAN      PStateField = 0b000_100
    SPSel    PStateField = 0b000_101
    SSBS     PStateField = 0b011_001
    DIT      PStateField = 0b011_010
    TCO      PStateField = 0b011_100
    DAIFSet  PStateField = 0b011_110
    DAIFClr  PStateField = 0b011_111
)

type (
    ATOption    uint16  // ATOption represents one of the Address Translation options.
    BRBOption   uint16  // BRBOption represents one of the Branch Record Buffer options.
    DCOption    uint16  // DCOption represents one of the Data Cache options.
    ICOption    uint16  // ICOption represents one of the Instruction Cache options.
    TLBIOption  uint16  // TLBIOption represents one of the TLB Invalidation (Pair) options.
)

const (
    S1E1R  ATOption = 0b000_1000_000
    S1E1W  ATOption = 0b000_1000_001
    S1E0R  ATOption = 0b000_1000_010
    S1E0W  ATOption = 0b000_1000_011
    S1E1RP ATOption = 0b000_1001_000
    S1E1WP ATOption = 0b000_1001_001
    S1E2R  ATOption = 0b100_1000_000
    S1E2W  ATOption = 0b100_1000_001
    S12E1R ATOption = 0b100_1000_100
    S12E1W ATOption = 0b100_1000_101
    S12E0R ATOption = 0b100_1000_110
    S12E0W ATOption = 0b100_1000_111
    S1E3R  ATOption = 0b110_1000_000
    S1E3W  ATOption = 0b110_1000_001
)

const (
    IALL BRBOption = 0b100
    INJ  BRBOption = 0b101
)

const (
    IVAC     DCOption = 0b000_0110_001
    ISW      DCOption = 0b000_0110_010
    IGVAC    DCOption = 0b000_0110_011
    IGSW     DCOption = 0b000_0110_100
    IGDVAC   DCOption = 0b000_0110_101
    IGDSW    DCOption = 0b000_0110_110
    CSW      DCOption = 0b000_1010_010
    CGSW     DCOption = 0b000_1010_100
    CGDSW    DCOption = 0b000_1010_110
    CISW     DCOption = 0b000_1110_010
    CIGSW    DCOption = 0b000_1110_100
    CIGDSW   DCOption = 0b000_1110_110
    ZVA      DCOption = 0b011_0100_001
    GVA      DCOption = 0b011_0100_011
    GZVA     DCOption = 0b011_0100_100
    CVAC     DCOption = 0b011_1010_001
    CGVAC    DCOption = 0b011_1010_011
    CGDVAC   DCOption = 0b011_1010_101
    CVAU     DCOption = 0b011_1011_001
    CVAP     DCOption = 0b011_1100_001
    CGVAP    DCOption = 0b011_1100_011
    CGDVAP   DCOption = 0b011_1100_101
    CVADP    DCOption = 0b011_1101_001
    CGVADP   DCOption = 0b011_1101_011
    CGDVADP  DCOption = 0b011_1101_101
    CIVAC    DCOption = 0b011_1110_001
    CIGVAC   DCOption = 0b011_1110_011
    CIGDVAC  DCOption = 0b011_1110_101
    CIPAE    DCOption = 0b100_1110_000
    CIGDPAE  DCOption = 0b100_1110_111
    CIPAPA   DCOption = 0b110_1110_001
    CIGDPAPA DCOption = 0b110_1110_101
)

const (
    IALLUIS ICOption = 0b000_0001_000
    IALLU   ICOption = 0b000_0101_000
    IVAU    ICOption = 0b011_0101_001
)

const (
    VMALLE1OS       TLBIOption = 0b000_1000_0001_000
    VAE1OS          TLBIOption = 0b000_1000_0001_001
    ASIDE1OS        TLBIOption = 0b000_1000_0001_010
    VAAE1OS         TLBIOption = 0b000_1000_0001_011
    VALE1OS         TLBIOption = 0b000_1000_0001_101
    VAALE1OS        TLBIOption = 0b000_1000_0001_111
    RVAE1IS         TLBIOption = 0b000_1000_0010_001
    RVAAE1IS        TLBIOption = 0b000_1000_0010_011
    RVALE1IS        TLBIOption = 0b000_1000_0010_101
    RVAALE1IS       TLBIOption = 0b000_1000_0010_111
    VMALLE1IS       TLBIOption = 0b000_1000_0011_000
    VAE1IS          TLBIOption = 0b000_1000_0011_001
    ASIDE1IS        TLBIOption = 0b000_1000_0011_010
    VAAE1IS         TLBIOption = 0b000_1000_0011_011
    VALE1IS         TLBIOption = 0b000_1000_0011_101
    VAALE1IS        TLBIOption = 0b000_1000_0011_111
    RVAE1OS         TLBIOption = 0b000_1000_0101_001
    RVAAE1OS        TLBIOption = 0b000_1000_0101_011
    RVALE1OS        TLBIOption = 0b000_1000_0101_101
    RVAALE1OS       TLBIOption = 0b000_1000_0101_111
    RVAE1           TLBIOption = 0b000_1000_0110_001
    RVAAE1          TLBIOption = 0b000_1000_0110_011
    RVALE1          TLBIOption = 0b000_1000_0110_101
    RVAALE1         TLBIOption = 0b000_1000_0110_111
    VMALLE1         TLBIOption = 0b000_1000_0111_000
    VAE1            TLBIOption = 0b000_1000_0111_001
    ASIDE1          TLBIOption = 0b000_1000_0111_010
    VAAE1           TLBIOption = 0b000_1000_0111_011
    VALE1           TLBIOption = 0b000_1000_0111_101
    VAALE1          TLBIOption = 0b000_1000_0111_111
    VMALLE1OSNXS    TLBIOption = 0b000_1001_0001_000
    VAE1OSNXS       TLBIOption = 0b000_1001_0001_001
    ASIDE1OSNXS     TLBIOption = 0b000_1001_0001_010
    VAAE1OSNXS      TLBIOption = 0b000_1001_0001_011
    VALE1OSNXS      TLBIOption = 0b000_1001_0001_101
    VAALE1OSNXS     TLBIOption = 0b000_1001_0001_111
    RVAE1ISNXS      TLBIOption = 0b000_1001_0010_001
    RVAAE1ISNXS     TLBIOption = 0b000_1001_0010_011
    RVALE1ISNXS     TLBIOption = 0b000_1001_0010_101
    RVAALE1ISNXS    TLBIOption = 0b000_1001_0010_111
    VMALLE1ISNXS    TLBIOption = 0b000_1001_0011_000
    VAE1ISNXS       TLBIOption = 0b000_1001_0011_001
    ASIDE1ISNXS     TLBIOption = 0b000_1001_0011_010
    VAAE1ISNXS      TLBIOption = 0b000_1001_0011_011
    VALE1ISNXS      TLBIOption = 0b000_1001_0011_101
    VAALE1ISNXS     TLBIOption = 0b000_1001_0011_111
    RVAE1OSNXS      TLBIOption = 0b000_1001_0101_001
    RVAAE1OSNXS     TLBIOption = 0b000_1001_0101_011
    RVALE1OSNXS     TLBIOption = 0b000_1001_0101_101
    RVAALE1OSNXS    TLBIOption = 0b000_1001_0101_111
    RVAE1NXS        TLBIOption = 0b000_1001_0110_001
    RVAAE1NXS       TLBIOption = 0b000_1001_0110_011
    RVALE1NXS       TLBIOption = 0b000_1001_0110_101
    RVAALE1NXS      TLBIOption = 0b000_1001_0110_111
    VMALLE1NXS      TLBIOption = 0b000_1001_0111_000
    VAE1NXS         TLBIOption = 0b000_1001_0111_001
    ASIDE1NXS       TLBIOption = 0b000_1001_0111_010
    VAAE1NXS        TLBIOption = 0b000_1001_0111_011
    VALE1NXS        TLBIOption = 0b000_1001_0111_101
    VAALE1NXS       TLBIOption = 0b000_1001_0111_111
    IPAS2E1IS       TLBIOption = 0b100_1000_0000_001
    RIPAS2E1IS      TLBIOption = 0b100_1000_0000_010
    IPAS2LE1IS      TLBIOption = 0b100_1000_0000_101
    RIPAS2LE1IS     TLBIOption = 0b100_1000_0000_110
    ALLE2OS         TLBIOption = 0b100_1000_0001_000
    VAE2OS          TLBIOption = 0b100_1000_0001_001
    ALLE1OS         TLBIOption = 0b100_1000_0001_100
    VALE2OS         TLBIOption = 0b100_1000_0001_101
    VMALLS12E1OS    TLBIOption = 0b100_1000_0001_110
    RVAE2IS         TLBIOption = 0b100_1000_0010_001
    RVALE2IS        TLBIOption = 0b100_1000_0010_101
    ALLE2IS         TLBIOption = 0b100_1000_0011_000
    VAE2IS          TLBIOption = 0b100_1000_0011_001
    ALLE1IS         TLBIOption = 0b100_1000_0011_100
    VALE2IS         TLBIOption = 0b100_1000_0011_101
    VMALLS12E1IS    TLBIOption = 0b100_1000_0011_110
    IPAS2E1OS       TLBIOption = 0b100_1000_0100_000
    IPAS2E1         TLBIOption = 0b100_1000_0100_001
    RIPAS2E1        TLBIOption = 0b100_1000_0100_010
    RIPAS2E1OS      TLBIOption = 0b100_1000_0100_011
    IPAS2LE1OS      TLBIOption = 0b100_1000_0100_100
    IPAS2LE1        TLBIOption = 0b100_1000_0100_101
    RIPAS2LE1       TLBIOption = 0b100_1000_0100_110
    RIPAS2LE1OS     TLBIOption = 0b100_1000_0100_111
    RVAE2OS         TLBIOption = 0b100_1000_0101_001
    RVALE2OS        TLBIOption = 0b100_1000_0101_101
    RVAE2           TLBIOption = 0b100_1000_0110_001
    RVALE2          TLBIOption = 0b100_1000_0110_101
    ALLE2           TLBIOption = 0b100_1000_0111_000
    VAE2            TLBIOption = 0b100_1000_0111_001
    ALLE1           TLBIOption = 0b100_1000_0111_100
    VALE2           TLBIOption = 0b100_1000_0111_101
    VMALLS12E1      TLBIOption = 0b100_1000_0111_110
    IPAS2E1ISNXS    TLBIOption = 0b100_1001_0000_001
    RIPAS2E1ISNXS   TLBIOption = 0b100_1001_0000_010
    IPAS2LE1ISNXS   TLBIOption = 0b100_1001_0000_101
    RIPAS2LE1ISNXS  TLBIOption = 0b100_1001_0000_110
    ALLE2OSNXS      TLBIOption = 0b100_1001_0001_000
    VAE2OSNXS       TLBIOption = 0b100_1001_0001_001
    ALLE1OSNXS      TLBIOption = 0b100_1001_0001_100
    VALE2OSNXS      TLBIOption = 0b100_1001_0001_101
    VMALLS12E1OSNXS TLBIOption = 0b100_1001_0001_110
    RVAE2ISNXS      TLBIOption = 0b100_1001_0010_001
    RVALE2ISNXS     TLBIOption = 0b100_1001_0010_101
    ALLE2ISNXS      TLBIOption = 0b100_1001_0011_000
    VAE2ISNXS       TLBIOption = 0b100_1001_0011_001
    ALLE1ISNXS      TLBIOption = 0b100_1001_0011_100
    VALE2ISNXS      TLBIOption = 0b100_1001_0011_101
    VMALLS12E1ISNXS TLBIOption = 0b100_1001_0011_110
    IPAS2E1OSNXS    TLBIOption = 0b100_1001_0100_000
    IPAS2E1NXS      TLBIOption = 0b100_1001_0100_001
    RIPAS2E1NXS     TLBIOption = 0b100_1001_0100_010
    RIPAS2E1OSNXS   TLBIOption = 0b100_1001_0100_011
    IPAS2LE1OSNXS   TLBIOption = 0b100_1001_0100_100
    IPAS2LE1NXS     TLBIOption = 0b100_1001_0100_101
    RIPAS2LE1NXS    TLBIOption = 0b100_1001_0100_110
    RIPAS2LE1OSNXS  TLBIOption = 0b100_1001_0100_111
    RVAE2OSNXS      TLBIOption = 0b100_1001_0101_001
    RVALE2OSNXS     TLBIOption = 0b100_1001_0101_101
    RVAE2NXS        TLBIOption = 0b100_1001_0110_001
    RVALE2NXS       TLBIOption = 0b100_1001_0110_101
    ALLE2NXS        TLBIOption = 0b100_1001_0111_000
    VAE2NXS         TLBIOption = 0b100_1001_0111_001
    ALLE1NXS        TLBIOption = 0b100_1001_0111_100
    VALE2NXS        TLBIOption = 0b100_1001_0111_101
    VMALLS12E1NXS   TLBIOption = 0b100_1001_0111_110
    ALLE3OS         TLBIOption = 0b110_1000_0001_000
    VAE3OS          TLBIOption = 0b110_1000_0001_001
    PAALLOS         TLBIOption = 0b110_1000_0001_100
    VALE3OS         TLBIOption = 0b110_1000_0001_101
    RVAE3IS         TLBIOption = 0b110_1000_0010_001
    RVALE3IS        TLBIOption = 0b110_1000_0010_101
    ALLE3IS         TLBIOption = 0b110_1000_0011_000
    VAE3IS          TLBIOption = 0b110_1000_0011_001
    VALE3IS         TLBIOption = 0b110_1000_0011_101
    RPAOS           TLBIOption = 0b110_1000_0100_011
    RPALOS          TLBIOption = 0b110_1000_0100_111
    RVAE3OS         TLBIOption = 0b110_1000_0101_001
    RVALE3OS        TLBIOption = 0b110_1000_0101_101
    RVAE3           TLBIOption = 0b110_1000_0110_001
    RVALE3          TLBIOption = 0b110_1000_0110_101
    ALLE3           TLBIOption = 0b110_1000_0111_000
    VAE3            TLBIOption = 0b110_1000_0111_001
    PAALL           TLBIOption = 0b110_1000_0111_100
    VALE3           TLBIOption = 0b110_1000_0111_101
    ALLE3OSNXS      TLBIOption = 0b110_1001_0001_000
    VAE3OSNXS       TLBIOption = 0b110_1001_0001_001
    VALE3OSNXS      TLBIOption = 0b110_1001_0001_101
    RVAE3ISNXS      TLBIOption = 0b110_1001_0010_001
    RVALE3ISNXS     TLBIOption = 0b110_1001_0010_101
    ALLE3ISNXS      TLBIOption = 0b110_1001_0011_000
    VAE3ISNXS       TLBIOption = 0b110_1001_0011_001
    VALE3ISNXS      TLBIOption = 0b110_1001_0011_101
    RVAE3OSNXS      TLBIOption = 0b110_1001_0101_001
    RVALE3OSNXS     TLBIOption = 0b110_1001_0101_101
    RVAE3NXS        TLBIOption = 0b110_1001_0110_001
    RVALE3NXS       TLBIOption = 0b110_1001_0110_101
    ALLE3NXS        TLBIOption = 0b110_1001_0111_000
    VAE3NXS         TLBIOption = 0b110_1001_0111_001
    VALE3NXS        TLBIOption = 0b110_1001_0111_101
)

// SMEOption represents one of the options used by SMSTART / SMSTOP instructions.
type SMEOption uint8

const (
    SM SMEOption = 0b0011   // SMSTART SM enters Streaming SVE mode, but does not enable the SME ZA storage.
    ZA SMEOption = 0b0101   // SMSTART ZA enables the SME ZA storage, but does not cause an entry to Streaming SVE mode.
)

// Symbol represents one of the instruction literal symbol.
type Symbol uint8

const (
    CSYNC Symbol = iota
    DSYNC
    RCTX
)

// PrefetchOp represents one of the Prefetch Options.
type PrefetchOp uint8

const (
    PLD_L1_KEEP PrefetchOp = 0b00000
    PLD_L1_STRM PrefetchOp = 0b00001
    PLD_L2_KEEP PrefetchOp = 0b00010
    PLD_L2_STRM PrefetchOp = 0b00011
    PLD_L3_KEEP PrefetchOp = 0b00100
    PLD_L3_STRM PrefetchOp = 0b00101
    PLI_L1_KEEP PrefetchOp = 0b01000
    PLI_L1_STRM PrefetchOp = 0b01001
    PLI_L2_KEEP PrefetchOp = 0b01010
    PLI_L2_STRM PrefetchOp = 0b01011
    PLI_L3_KEEP PrefetchOp = 0b01100
    PLI_L3_STRM PrefetchOp = 0b01101
    PST_L1_KEEP PrefetchOp = 0b10000
    PST_L1_STRM PrefetchOp = 0b10001
    PST_L2_KEEP PrefetchOp = 0b10010
    PST_L2_STRM PrefetchOp = 0b10011
    PST_L3_KEEP PrefetchOp = 0b10100
    PST_L3_STRM PrefetchOp = 0b10101
)

// RangePrefetchOp represents one of the Range Prefetch Options.
type RangePrefetchOp uint8

const (
    PLD_KEEP RangePrefetchOp = 0b000000
    PLD_STRM RangePrefetchOp = 0b000100
    PST_KEEP RangePrefetchOp = 0b000001
    PST_STRM RangePrefetchOp = 0b000101
)

func (self RangePrefetchOp) encode() uint32 {
    s  := uint32(self & 0b001000)
    o0 := uint32(self & 0b010000)
    o2 := uint32(self & 0b100000)
    rt := uint32(self & 0b000111)
    return s | (rt << 4) | (o2 >> 3) | (o0 >> 4) | 0b110000010
}
