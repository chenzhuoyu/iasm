package aarch64

var _Modifiers = map[string]func(uint8)Modifier {
    "MSL"  : func(n uint8) Modifier { return MSL(n)  },
    "LSL"  : func(n uint8) Modifier { return LSL(n)  },
    "LSR"  : func(n uint8) Modifier { return LSR(n)  },
    "ASR"  : func(n uint8) Modifier { return ASR(n)  },
    "ROR"  : func(n uint8) Modifier { return ROR(n)  },
    "UXTB" : func(n uint8) Modifier { return UXTB(n) },
    "UXTH" : func(n uint8) Modifier { return UXTH(n) },
    "UXTW" : func(n uint8) Modifier { return UXTW(n) },
    "UXTX" : func(n uint8) Modifier { return UXTX(n) },
    "SXTB" : func(n uint8) Modifier { return SXTB(n) },
    "SXTH" : func(n uint8) Modifier { return SXTH(n) },
    "SXTW" : func(n uint8) Modifier { return SXTW(n) },
    "SXTX" : func(n uint8) Modifier { return SXTX(n) },
}
