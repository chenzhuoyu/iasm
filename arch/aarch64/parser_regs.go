package aarch64

import (
    `fmt`

    `github.com/chenzhuoyu/iasm/asm`
)

type (
	_VReg uint8
)

const (
    _V0 _VReg = iota
    _V1
    _V2
    _V3
    _V4
    _V5
    _V6
    _V7
    _V8
    _V9
    _V10
    _V11
    _V12
    _V13
    _V14
    _V15
    _V16
    _V17
    _V18
    _V19
    _V20
    _V21
    _V22
    _V23
    _V24
    _V25
    _V26
    _V27
    _V28
    _V29
    _V30
    _V31
)

var _VecFormats = map[string]VecFormat {
    "8B"  : Vec8B,
    "16B" : Vec16B,
    "2H"  : Vec2H,
    "4H"  : Vec4H,
    "8H"  : Vec8H,
    "2S"  : Vec2S,
    "4S"  : Vec4S,
    "1D"  : Vec1D,
    "2D"  : Vec2D,
    "1Q"  : Vec1Q,
}

var _VecIndexModes = map[string]VecIndexMode {
    "B"  : ModeB,
    "4B" : Mode4B,
    "H"  : ModeH,
    "2H" : Mode2H,
    "S"  : ModeS,
    "D"  : ModeD,
}

func (self _VReg) ID()     uint8  { return uint8(self) }
func (self _VReg) String() string { return fmt.Sprintf("v%d", self.ID()) }

func (self _VReg) idx(index uint8) uint16 {
    return (uint16(self) << 8) + uint16(index)
}

func (self _VReg) toVec(fmt VecFormat) VRegister {
    switch fmt {
        case Vec8B  : return VRegister8B  (self)
        case Vec16B : return VRegister16B (self)
        case Vec2H  : return VRegister2H  (self)
        case Vec4H  : return VRegister4H  (self)
        case Vec8H  : return VRegister8H  (self)
        case Vec2S  : return VRegister2S  (self)
        case Vec4S  : return VRegister4S  (self)
        case Vec1D  : return VRegister1D  (self)
        case Vec2D  : return VRegister2D  (self)
        case Vec1Q  : return VRegister1Q  (self)
        default     : panic("aarch64: unknown vector format")
    }
}

func (self _VReg) toIndex(mode VecIndexMode, index uint8) VidxRegister {
    switch mode {
        case ModeB  : return VidxRegisterB  (self.idx(index))
        case Mode4B : return VidxRegister4B (self.idx(index))
        case ModeH  : return VidxRegisterH  (self.idx(index))
        case Mode2H : return VidxRegister2H (self.idx(index))
        case ModeS  : return VidxRegisterS  (self.idx(index))
        case ModeD  : return VidxRegisterD  (self.idx(index))
        default     : panic("aarch64: unknown vector index mode")
    }
}

var _CoreRegisters = map[string]asm.Register {
    "W0"  : W0,
    "W1"  : W1,
    "W2"  : W2,
    "W3"  : W3,
    "W4"  : W4,
    "W5"  : W5,
    "W6"  : W6,
    "W7"  : W7,
    "W8"  : W8,
    "W9"  : W9,
    "W10" : W10,
    "W11" : W11,
    "W12" : W12,
    "W13" : W13,
    "W14" : W14,
    "W15" : W15,
    "W16" : W16,
    "W17" : W17,
    "W18" : W18,
    "W19" : W19,
    "W20" : W20,
    "W21" : W21,
    "W22" : W22,
    "W23" : W23,
    "W24" : W24,
    "W25" : W25,
    "W26" : W26,
    "W27" : W27,
    "W28" : W28,
    "W29" : W29,
    "W30" : W30,
    "WZR" : WZR,
    "WSP" : WSP,
    "X0"  : X0,
    "X1"  : X1,
    "X2"  : X2,
    "X3"  : X3,
    "X4"  : X4,
    "X5"  : X5,
    "X6"  : X6,
    "X7"  : X7,
    "X8"  : X8,
    "X9"  : X9,
    "X10" : X10,
    "X11" : X11,
    "X12" : X12,
    "X13" : X13,
    "X14" : X14,
    "X15" : X15,
    "X16" : X16,
    "X17" : X17,
    "X18" : X18,
    "X19" : X19,
    "X20" : X20,
    "X21" : X21,
    "X22" : X22,
    "X23" : X23,
    "X24" : X24,
    "X25" : X25,
    "X26" : X26,
    "X27" : X27,
    "X28" : X28,
    "X29" : X29,
    "X30" : X30,
    "XZR" : XZR,
    "SP"  : SP,
    "XR"  : XR,
    "IP0" : IP0,
    "IP1" : IP1,
    "PR"  : PR,
    "FP"  : FP,
    "LR"  : LR,
}

var _SimdRegisters = map[string]asm.Register {
    "B0"  : B0,
    "B1"  : B1,
    "B2"  : B2,
    "B3"  : B3,
    "B4"  : B4,
    "B5"  : B5,
    "B6"  : B6,
    "B7"  : B7,
    "B8"  : B8,
    "B9"  : B9,
    "B10" : B10,
    "B11" : B11,
    "B12" : B12,
    "B13" : B13,
    "B14" : B14,
    "B15" : B15,
    "B16" : B16,
    "B17" : B17,
    "B18" : B18,
    "B19" : B19,
    "B20" : B20,
    "B21" : B21,
    "B22" : B22,
    "B23" : B23,
    "B24" : B24,
    "B25" : B25,
    "B26" : B26,
    "B27" : B27,
    "B28" : B28,
    "B29" : B29,
    "B30" : B30,
    "B31" : B31,
    "H0"  : H0,
    "H1"  : H1,
    "H2"  : H2,
    "H3"  : H3,
    "H4"  : H4,
    "H5"  : H5,
    "H6"  : H6,
    "H7"  : H7,
    "H8"  : H8,
    "H9"  : H9,
    "H10" : H10,
    "H11" : H11,
    "H12" : H12,
    "H13" : H13,
    "H14" : H14,
    "H15" : H15,
    "H16" : H16,
    "H17" : H17,
    "H18" : H18,
    "H19" : H19,
    "H20" : H20,
    "H21" : H21,
    "H22" : H22,
    "H23" : H23,
    "H24" : H24,
    "H25" : H25,
    "H26" : H26,
    "H27" : H27,
    "H28" : H28,
    "H29" : H29,
    "H30" : H30,
    "H31" : H31,
    "S0"  : S0,
    "S1"  : S1,
    "S2"  : S2,
    "S3"  : S3,
    "S4"  : S4,
    "S5"  : S5,
    "S6"  : S6,
    "S7"  : S7,
    "S8"  : S8,
    "S9"  : S9,
    "S10" : S10,
    "S11" : S11,
    "S12" : S12,
    "S13" : S13,
    "S14" : S14,
    "S15" : S15,
    "S16" : S16,
    "S17" : S17,
    "S18" : S18,
    "S19" : S19,
    "S20" : S20,
    "S21" : S21,
    "S22" : S22,
    "S23" : S23,
    "S24" : S24,
    "S25" : S25,
    "S26" : S26,
    "S27" : S27,
    "S28" : S28,
    "S29" : S29,
    "S30" : S30,
    "S31" : S31,
    "D0"  : D0,
    "D1"  : D1,
    "D2"  : D2,
    "D3"  : D3,
    "D4"  : D4,
    "D5"  : D5,
    "D6"  : D6,
    "D7"  : D7,
    "D8"  : D8,
    "D9"  : D9,
    "D10" : D10,
    "D11" : D11,
    "D12" : D12,
    "D13" : D13,
    "D14" : D14,
    "D15" : D15,
    "D16" : D16,
    "D17" : D17,
    "D18" : D18,
    "D19" : D19,
    "D20" : D20,
    "D21" : D21,
    "D22" : D22,
    "D23" : D23,
    "D24" : D24,
    "D25" : D25,
    "D26" : D26,
    "D27" : D27,
    "D28" : D28,
    "D29" : D29,
    "D30" : D30,
    "D31" : D31,
    "Q0"  : Q0,
    "Q1"  : Q1,
    "Q2"  : Q2,
    "Q3"  : Q3,
    "Q4"  : Q4,
    "Q5"  : Q5,
    "Q6"  : Q6,
    "Q7"  : Q7,
    "Q8"  : Q8,
    "Q9"  : Q9,
    "Q10" : Q10,
    "Q11" : Q11,
    "Q12" : Q12,
    "Q13" : Q13,
    "Q14" : Q14,
    "Q15" : Q15,
    "Q16" : Q16,
    "Q17" : Q17,
    "Q18" : Q18,
    "Q19" : Q19,
    "Q20" : Q20,
    "Q21" : Q21,
    "Q22" : Q22,
    "Q23" : Q23,
    "Q24" : Q24,
    "Q25" : Q25,
    "Q26" : Q26,
    "Q27" : Q27,
    "Q28" : Q28,
    "Q29" : Q29,
    "Q30" : Q30,
    "Q31" : Q31,
}

var _VectorRegisters = map[string]_VReg {
    "V0"  : _V0,
    "V1"  : _V1,
    "V2"  : _V2,
    "V3"  : _V3,
    "V4"  : _V4,
    "V5"  : _V5,
    "V6"  : _V6,
    "V7"  : _V7,
    "V8"  : _V8,
    "V9"  : _V9,
    "V10" : _V10,
    "V11" : _V11,
    "V12" : _V12,
    "V13" : _V13,
    "V14" : _V14,
    "V15" : _V15,
    "V16" : _V16,
    "V17" : _V17,
    "V18" : _V18,
    "V19" : _V19,
    "V20" : _V20,
    "V21" : _V21,
    "V22" : _V22,
    "V23" : _V23,
    "V24" : _V24,
    "V25" : _V25,
    "V26" : _V26,
    "V27" : _V27,
    "V28" : _V28,
    "V29" : _V29,
    "V30" : _V30,
    "V31" : _V31,
}
