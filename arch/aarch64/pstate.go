package aarch64

// PStateField represents one of the processor states.
type PStateField uint8

const (
    _UAO     PStateField = 0b000_011_000
    _PAN     PStateField = 0b000_100_000
    _SPSel   PStateField = 0b000_101_000
    _ALLINT  PStateField = 0b001_000_000
    _PM      PStateField = 0b001_000_001
    _SSBS    PStateField = 0b011_001_000
    _DIT     PStateField = 0b011_010_000
    SVCRSM   PStateField = 0b011_011_001
    SVCRZA   PStateField = 0b011_011_010
    SVCRSMZA PStateField = 0b011_011_011
    _TCO     PStateField = 0b011_100_000
    DAIFSet  PStateField = 0b011_110_000
    DAIFClr  PStateField = 0b011_111_000
)

var _PStateMax = map[PStateField]uint64 {
    _UAO      : 15,
    _PAN      : 15,
    _SPSel    : 15,
    _ALLINT   : 1,
    _PM       : 1,
    _SSBS     : 15,
    _DIT      : 15,
    SVCRSM    : 1,
    SVCRZA    : 1,
    SVCRSMZA  : 1,
    _TCO      : 15,
    DAIFSet   : 15,
    DAIFClr   : 15,
}

var _PStateTab = map[string]PStateField {
    "SVCRSM"   : SVCRSM,
    "SVCRZA"   : SVCRZA,
    "SVCRSMZA" : SVCRSMZA,
    "DAIFSet"  : DAIFSet,
    "DAIFClr"  : DAIFClr,
}

var _PStateNames = map[PStateField]string {
    _UAO      : "UAO",
    _PAN      : "PAN",
    _SPSel    : "SPSel",
    _ALLINT   : "ALLINT",
    _PM       : "PM",
    _SSBS     : "SSBS",
    _DIT      : "DIT",
    SVCRSM    : "SVCRSM",
    SVCRZA    : "SVCRZA",
    SVCRSMZA  : "SVCRSMZA",
    _TCO      : "TCO",
    DAIFSet   : "DAIFSet",
    DAIFClr   : "DAIFClr",
}

var _SysRegPStateMap = map[SystemRegister]PStateField {
    UAO    : _UAO,
    PAN    : _PAN,
    SPSel  : _SPSel,
    ALLINT : _ALLINT,
    PM     : _PM,
    SSBS   : _SSBS,
    DIT    : _DIT,
    TCO    : _TCO,
}

func (self PStateField) ID() uint8 {
    panic("aarch64: invalid use of Processor State Fields.")
}

func (self PStateField) String() string {
    if ret, ok := _PStateNames[self]; ok {
        return ret
    } else {
        return "???"
    }
}

func (self PStateField) encode() uint32 {
    return uint32(self) << 1
}
