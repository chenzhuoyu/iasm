package aarch64

import (
    `fmt`
)

// Register represents a generic hardware register.
type Register interface {
    fmt.Stringer
    ID() uint8
}

type (
    Register32 uint8
    Register64 uint8
)

func (self Register32) ID() uint8 {
    switch {
        case self == WSP          : return uint8(WZR)
        case self &^ 0b11111 == 0 : return uint8(self)
        default                   : panic("aarch64: invalid generic 32-bit register")
    }
}

func (self Register64) ID() uint8 {
    switch {
        case self == SP           : return uint8(XZR)
        case self &^ 0b11111 == 0 : return uint8(self)
        default                   : panic("aarch64: invalid generic 64-bit register")
    }
}

func (self Register32) String() string {
    switch self {
        case WSP : return "wsp"
        case WZR : return "wzr"
        default  : return fmt.Sprintf("w%d", self.ID())
    }
}

func (self Register64) String() string {
    switch self {
        case LR  : return "lr"
        case SP  : return "sp"
        case XZR : return "xzr"
        default  : return fmt.Sprintf("x%d", self.ID())
    }
}

const (
    W0 Register32 = iota
    W1
    W2
    W3
    W4
    W5
    W6
    W7
    W8
    W9
    W10
    W11
    W12
    W13
    W14
    W15
    W16
    W17
    W18
    W19
    W20
    W21
    W22
    W23
    W24
    W25
    W26
    W27
    W28
    W29
    W30
    WZR
)

const (
    X0 Register64 = iota
    X1
    X2
    X3
    X4
    X5
    X6
    X7
    X8
    X9
    X10
    X11
    X12
    X13
    X14
    X15
    X16
    X17
    X18
    X19
    X20
    X21
    X22
    X23
    X24
    X25
    X26
    X27
    X28
    X29
    X30
    XZR
)

const (
    LR  = X30
    SP  = XZR | 0x80
    WSP = WZR | 0x80
)
