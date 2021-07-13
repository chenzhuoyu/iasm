package x86_64

// Register represents a hardware register.
type Register interface {
    implRegister()
}

type (
    Register8  byte
    Register16 byte
    Register32 byte
    Register64 byte
)

type (
    KRegister   byte
    MMRegister  byte
    XMMRegister byte
    YMMRegister byte
    ZMMRegister byte
)

// RegisterMask is a KRegister used to mask another register.
type RegisterMask struct {
    Z bool
    K KRegister
}

// MaskedRegister is a Register masked by a RegisterMask.
type MaskedRegister struct {
    Reg  Register
    Mask RegisterMask
}

const (
    AL Register8 = iota
    CL
    DL
    BL
    SPL
    BPL
    SIL
    DIL
    R8b
    R9b
    R10b
    R11b
    R12b
    R13b
    R14b
    R15b
)

const (
    AH = SPL | 0x80
    CH = BPL | 0x80
    DH = SIL | 0x80
    BH = DIL | 0x80
)

const (
    AX Register16 = iota
    CX
    DX
    BX
    SP
    BP
    SI
    DI
    R8w
    R9w
    R10w
    R11w
    R12w
    R13w
    R14w
    R15w
)

const (
    EAX Register32 = iota
    ECX
    EDX
    EBX
    ESP
    EBP
    ESI
    EDI
    R8d
    R9d
    R10d
    R11d
    R12d
    R13d
    R14d
    R15d
)

const (
    RAX Register64 = iota
    RCX
    RDX
    RBX
    RSP
    RBP
    RSI
    RDI
    R8
    R9
    R10
    R11
    R12
    R13
    R14
    R15
)

const (
    K0 KRegister = iota
    K1
    K2
    K3
    K4
    K5
    K6
    K7
)

const (
    MM0 MMRegister = iota
    MM1
    MM2
    MM3
    MM4
    MM5
    MM6
    MM7
)

const (
    XMM0 XMMRegister = iota
    XMM1
    XMM2
    XMM3
    XMM4
    XMM5
    XMM6
    XMM7
    XMM8
    XMM9
    XMM10
    XMM11
    XMM12
    XMM13
    XMM14
    XMM15
    XMM16
    XMM17
    XMM18
    XMM19
    XMM20
    XMM21
    XMM22
    XMM23
    XMM24
    XMM25
    XMM26
    XMM27
    XMM28
    XMM29
    XMM30
    XMM31
)

const (
    YMM0 YMMRegister = iota
    YMM1
    YMM2
    YMM3
    YMM4
    YMM5
    YMM6
    YMM7
    YMM8
    YMM9
    YMM10
    YMM11
    YMM12
    YMM13
    YMM14
    YMM15
    YMM16
    YMM17
    YMM18
    YMM19
    YMM20
    YMM21
    YMM22
    YMM23
    YMM24
    YMM25
    YMM26
    YMM27
    YMM28
    YMM29
    YMM30
    YMM31
)

const (
    ZMM0 ZMMRegister = iota
    ZMM1
    ZMM2
    ZMM3
    ZMM4
    ZMM5
    ZMM6
    ZMM7
    ZMM8
    ZMM9
    ZMM10
    ZMM11
    ZMM12
    ZMM13
    ZMM14
    ZMM15
    ZMM16
    ZMM17
    ZMM18
    ZMM19
    ZMM20
    ZMM21
    ZMM22
    ZMM23
    ZMM24
    ZMM25
    ZMM26
    ZMM27
    ZMM28
    ZMM29
    ZMM30
    ZMM31
)

func (self Register8)  implRegister() {}
func (self Register16) implRegister() {}
func (self Register32) implRegister() {}
func (self Register64) implRegister() {}

func (self KRegister)   implRegister() {}
func (self MMRegister)  implRegister() {}
func (self XMMRegister) implRegister() {}
func (self YMMRegister) implRegister() {}
func (self ZMMRegister) implRegister() {}

// Registers maps register name into Register instances.
var Registers = map[string]Register {
    "al"    : AL,
    "cl"    : CL,
    "dl"    : DL,
    "bl"    : BL,
    "spl"   : SPL,
    "bpl"   : BPL,
    "sil"   : SIL,
    "dil"   : DIL,
    "r8b"   : R8b,
    "r9b"   : R9b,
    "r10b"  : R10b,
    "r11b"  : R11b,
    "r12b"  : R12b,
    "r13b"  : R13b,
    "r14b"  : R14b,
    "r15b"  : R15b,
    "ah"    : AH,
    "ch"    : CH,
    "dh"    : DH,
    "bh"    : BH,
    "ax"    : AX,
    "cx"    : CX,
    "dx"    : DX,
    "bx"    : BX,
    "sp"    : SP,
    "bp"    : BP,
    "si"    : SI,
    "di"    : DI,
    "r8w"   : R8w,
    "r9w"   : R9w,
    "r10w"  : R10w,
    "r11w"  : R11w,
    "r12w"  : R12w,
    "r13w"  : R13w,
    "r14w"  : R14w,
    "r15w"  : R15w,
    "eax"   : EAX,
    "ecx"   : ECX,
    "edx"   : EDX,
    "ebx"   : EBX,
    "esp"   : ESP,
    "ebp"   : EBP,
    "esi"   : ESI,
    "edi"   : EDI,
    "r8d"   : R8d,
    "r9d"   : R9d,
    "r10d"  : R10d,
    "r11d"  : R11d,
    "r12d"  : R12d,
    "r13d"  : R13d,
    "r14d"  : R14d,
    "r15d"  : R15d,
    "rax"   : RAX,
    "rcx"   : RCX,
    "rdx"   : RDX,
    "rbx"   : RBX,
    "rsp"   : RSP,
    "rbp"   : RBP,
    "rsi"   : RSI,
    "rdi"   : RDI,
    "r8"    : R8,
    "r9"    : R9,
    "r10"   : R10,
    "r11"   : R11,
    "r12"   : R12,
    "r13"   : R13,
    "r14"   : R14,
    "r15"   : R15,
    "k0"    : K0,
    "k1"    : K1,
    "k2"    : K2,
    "k3"    : K3,
    "k4"    : K4,
    "k5"    : K5,
    "k6"    : K6,
    "k7"    : K7,
    "mm0"   : MM0,
    "mm1"   : MM1,
    "mm2"   : MM2,
    "mm3"   : MM3,
    "mm4"   : MM4,
    "mm5"   : MM5,
    "mm6"   : MM6,
    "mm7"   : MM7,
    "xmm0"  : XMM0,
    "xmm1"  : XMM1,
    "xmm2"  : XMM2,
    "xmm3"  : XMM3,
    "xmm4"  : XMM4,
    "xmm5"  : XMM5,
    "xmm6"  : XMM6,
    "xmm7"  : XMM7,
    "xmm8"  : XMM8,
    "xmm9"  : XMM9,
    "xmm10" : XMM10,
    "xmm11" : XMM11,
    "xmm12" : XMM12,
    "xmm13" : XMM13,
    "xmm14" : XMM14,
    "xmm15" : XMM15,
    "xmm16" : XMM16,
    "xmm17" : XMM17,
    "xmm18" : XMM18,
    "xmm19" : XMM19,
    "xmm20" : XMM20,
    "xmm21" : XMM21,
    "xmm22" : XMM22,
    "xmm23" : XMM23,
    "xmm24" : XMM24,
    "xmm25" : XMM25,
    "xmm26" : XMM26,
    "xmm27" : XMM27,
    "xmm28" : XMM28,
    "xmm29" : XMM29,
    "xmm30" : XMM30,
    "xmm31" : XMM31,
    "ymm0"  : YMM0,
    "ymm1"  : YMM1,
    "ymm2"  : YMM2,
    "ymm3"  : YMM3,
    "ymm4"  : YMM4,
    "ymm5"  : YMM5,
    "ymm6"  : YMM6,
    "ymm7"  : YMM7,
    "ymm8"  : YMM8,
    "ymm9"  : YMM9,
    "ymm10" : YMM10,
    "ymm11" : YMM11,
    "ymm12" : YMM12,
    "ymm13" : YMM13,
    "ymm14" : YMM14,
    "ymm15" : YMM15,
    "ymm16" : YMM16,
    "ymm17" : YMM17,
    "ymm18" : YMM18,
    "ymm19" : YMM19,
    "ymm20" : YMM20,
    "ymm21" : YMM21,
    "ymm22" : YMM22,
    "ymm23" : YMM23,
    "ymm24" : YMM24,
    "ymm25" : YMM25,
    "ymm26" : YMM26,
    "ymm27" : YMM27,
    "ymm28" : YMM28,
    "ymm29" : YMM29,
    "ymm30" : YMM30,
    "ymm31" : YMM31,
    "zmm0"  : ZMM0,
    "zmm1"  : ZMM1,
    "zmm2"  : ZMM2,
    "zmm3"  : ZMM3,
    "zmm4"  : ZMM4,
    "zmm5"  : ZMM5,
    "zmm6"  : ZMM6,
    "zmm7"  : ZMM7,
    "zmm8"  : ZMM8,
    "zmm9"  : ZMM9,
    "zmm10" : ZMM10,
    "zmm11" : ZMM11,
    "zmm12" : ZMM12,
    "zmm13" : ZMM13,
    "zmm14" : ZMM14,
    "zmm15" : ZMM15,
    "zmm16" : ZMM16,
    "zmm17" : ZMM17,
    "zmm18" : ZMM18,
    "zmm19" : ZMM19,
    "zmm20" : ZMM20,
    "zmm21" : ZMM21,
    "zmm22" : ZMM22,
    "zmm23" : ZMM23,
    "zmm24" : ZMM24,
    "zmm25" : ZMM25,
    "zmm26" : ZMM26,
    "zmm27" : ZMM27,
    "zmm28" : ZMM28,
    "zmm29" : ZMM29,
    "zmm30" : ZMM30,
    "zmm31" : ZMM31,
}
