package aarch64

import (
    `fmt`
)

type (
    WRegister uint8 // 32-bit generic purpose registers (W0 ~ W30) and WZR / WSP.
    XRegister uint8 // 64-bit generic purpose registers (X0 ~ X30) and XZR / SP.
)

func (self WRegister) ID() uint8 {
    if self == WSP {
        return uint8(WZR)
    } else {
        return uint8(self)
    }
}

func (self XRegister) ID() uint8 {
    if self == SP {
        return uint8(XZR)
    } else {
        return uint8(self)
    }
}

func (self WRegister) String() string {
    switch self {
        case WSP : return "wsp"
        case WZR : return "wzr"
        default  : return fmt.Sprintf("w%d", self.ID())
    }
}

func (self XRegister) String() string {
    switch self {
        case LR  : return "lr"
        case SP  : return "sp"
        case XZR : return "xzr"
        default  : return fmt.Sprintf("x%d", self.ID())
    }
}

const (
    W0 WRegister = iota
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
    X0 XRegister = iota
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
    XR  = X8
    IP0 = X16
    IP1 = X17
    PR  = X18
    FP  = X29
    LR  = X30
    SP  = XZR | 0x80
    WSP = WZR | 0x80
)
