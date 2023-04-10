package aarch64

import (
    `fmt`

    `github.com/chenzhuoyu/iasm/asm`
)

type (
    BRegister uint8         // SIMD register with sixteen 8-bit elements (B0 ~ B31).
    HRegister uint8         // SIMD register with eight 16 bit elements (H0 ~ H31).
    SRegister uint8         // SIMD register with four 32-bit elements (S0 ~ S31).
    DRegister uint8         // SIMD register with two 64-bit elements (D0 ~ D31).
    QRegister uint8         // SIMD register with a single 128-bit value (Q0 ~ Q31).
)

func (self BRegister) ID() uint8 { return uint8(self) }
func (self HRegister) ID() uint8 { return uint8(self) }
func (self SRegister) ID() uint8 { return uint8(self) }
func (self DRegister) ID() uint8 { return uint8(self) }
func (self QRegister) ID() uint8 { return uint8(self) }

func (self BRegister) String() string { return fmt.Sprintf("b%d", self.ID()) }
func (self HRegister) String() string { return fmt.Sprintf("h%d", self.ID()) }
func (self SRegister) String() string { return fmt.Sprintf("s%d", self.ID()) }
func (self DRegister) String() string { return fmt.Sprintf("d%d", self.ID()) }
func (self QRegister) String() string { return fmt.Sprintf("q%d", self.ID()) }

// VecFormat represents one of the vector arrangement.
type VecFormat uint8

const (
    Vec8B VecFormat = iota
    Vec16B
    Vec2H
    Vec4H
    Vec8H
    Vec2S
    Vec4S
    Vec1D
    Vec2D
    Vec1Q
)

var _VecFormatNames = [...]string {
    "8B",
    "16B",
    "2H",
    "4H",
    "8H",
    "2S",
    "4S",
    "1D",
    "2D",
    "1Q",
}

func (self VecFormat) String() string {
    if int(self) >= len(_VecFormatNames) {
        return "???"
    } else {
        return _VecFormatNames[self]
    }
}

// VRegister represents a vector register with certain arrangement.
type VRegister interface {
    asm.Register
    Format() VecFormat
    VRegister()
}

type (
    VRegister8B  uint8      // SIMD register with arrangement 8B (V8b0 ~ V8b31).
    VRegister16B uint8      // SIMD register with arrangement 16B (V16b0 ~ V16b31).
    VRegister2H  uint8      // SIMD register with arrangement 2H (V2h0 ~ V2h31).
    VRegister4H  uint8      // SIMD register with arrangement 4H (V4h0 ~ V4h31).
    VRegister8H  uint8      // SIMD register with arrangement 8H (V8h0 ~ V8h31).
    VRegister2S  uint8      // SIMD register with arrangement 2S (V2s0 ~ V2s31).
    VRegister4S  uint8      // SIMD register with arrangement 4S (V4s0 ~ V4s31).
    VRegister1D  uint8      // SIMD register with arrangement 1D (V1d0 ~ V1d31).
    VRegister2D  uint8      // SIMD register with arrangement 2D (V2d0 ~ V2d31).
    VRegister1Q  uint8      // SIMD register with arrangement 1Q (V1q0 ~ V1q31).
)

func (VRegister8B)  Format() VecFormat { return Vec8B }
func (VRegister16B) Format() VecFormat { return Vec16B }
func (VRegister2H)  Format() VecFormat { return Vec2H }
func (VRegister4H)  Format() VecFormat { return Vec4H }
func (VRegister8H)  Format() VecFormat { return Vec8H }
func (VRegister2S)  Format() VecFormat { return Vec2S }
func (VRegister4S)  Format() VecFormat { return Vec4S }
func (VRegister1D)  Format() VecFormat { return Vec1D }
func (VRegister2D)  Format() VecFormat { return Vec2D }
func (VRegister1Q)  Format() VecFormat { return Vec1Q }

func (VRegister8B)  VRegister() {}
func (VRegister16B) VRegister() {}
func (VRegister2H)  VRegister() {}
func (VRegister4H)  VRegister() {}
func (VRegister8H)  VRegister() {}
func (VRegister2S)  VRegister() {}
func (VRegister4S)  VRegister() {}
func (VRegister1D)  VRegister() {}
func (VRegister2D)  VRegister() {}
func (VRegister1Q)  VRegister() {}

func (self VRegister8B)  ID() uint8 { return uint8(self) }
func (self VRegister16B) ID() uint8 { return uint8(self) }
func (self VRegister2H)  ID() uint8 { return uint8(self) }
func (self VRegister4H)  ID() uint8 { return uint8(self) }
func (self VRegister8H)  ID() uint8 { return uint8(self) }
func (self VRegister2S)  ID() uint8 { return uint8(self) }
func (self VRegister4S)  ID() uint8 { return uint8(self) }
func (self VRegister1D)  ID() uint8 { return uint8(self) }
func (self VRegister2D)  ID() uint8 { return uint8(self) }
func (self VRegister1Q)  ID() uint8 { return uint8(self) }

func (self VRegister8B)  String() string { return fmt.Sprintf("v%d.8B" , self.ID()) }
func (self VRegister16B) String() string { return fmt.Sprintf("v%d.16B", self.ID()) }
func (self VRegister2H)  String() string { return fmt.Sprintf("v%d.2H" , self.ID()) }
func (self VRegister4H)  String() string { return fmt.Sprintf("v%d.4H" , self.ID()) }
func (self VRegister8H)  String() string { return fmt.Sprintf("v%d.8H" , self.ID()) }
func (self VRegister2S)  String() string { return fmt.Sprintf("v%d.2S" , self.ID()) }
func (self VRegister4S)  String() string { return fmt.Sprintf("v%d.4S" , self.ID()) }
func (self VRegister1D)  String() string { return fmt.Sprintf("v%d.1D" , self.ID()) }
func (self VRegister2D)  String() string { return fmt.Sprintf("v%d.2D" , self.ID()) }
func (self VRegister1Q)  String() string { return fmt.Sprintf("v%d.1Q" , self.ID()) }

// VecIndexMode represents the unit size of an indexed vector register.
type VecIndexMode uint8

const (
    ModeB VecIndexMode = iota
    Mode4B
    ModeH
    Mode2H
    ModeS
    ModeD
)

var _VecIndexUnitNames = [...]string {
    "B",
    "4B",
    "H",
    "2H",
    "S",
    "D",
}

func (self VecIndexMode) String() string {
    if int(self) >= len(_VecIndexUnitNames) {
        return "???"
    } else {
        return _VecIndexUnitNames[self]
    }
}

// VidxRegister represents an indexed vector register.
type VidxRegister interface {
    asm.Register
    Index() uint8
    IndexMode() VecIndexMode
    VidxRegister()
}

type (
	VidxRegisterB  uint16   // Indexed SIMD register with unit size B (Vib0 ~ Vib31).
	VidxRegister4B uint16   // Indexed SIMD register with unit size 4B (Vi4b0 ~ Vi4b31).
	VidxRegisterH  uint16   // Indexed SIMD register with unit size H (Vih0 ~ Vih31).
	VidxRegister2H uint16   // Indexed SIMD register with unit size 2H (Vi2h0 ~ Vi2h31).
	VidxRegisterS  uint16   // Indexed SIMD register with unit size S (Vis0 ~ Vis31).
	VidxRegisterD  uint16   // Indexed SIMD register with unit size H (Vid0 ~ Vid31).
)

func (VidxRegisterB)  IndexMode() VecIndexMode { return ModeB  }
func (VidxRegister4B) IndexMode() VecIndexMode { return Mode4B }
func (VidxRegisterH)  IndexMode() VecIndexMode { return ModeH  }
func (VidxRegister2H) IndexMode() VecIndexMode { return Mode2H }
func (VidxRegisterS)  IndexMode() VecIndexMode { return ModeS  }
func (VidxRegisterD)  IndexMode() VecIndexMode { return ModeD  }

func (VidxRegisterB)  VidxRegister() {}
func (VidxRegister4B) VidxRegister() {}
func (VidxRegisterH)  VidxRegister() {}
func (VidxRegister2H) VidxRegister() {}
func (VidxRegisterS)  VidxRegister() {}
func (VidxRegisterD)  VidxRegister() {}

func (self VidxRegisterB)  ID() uint8 { return uint8(self >> 8) }
func (self VidxRegister4B) ID() uint8 { return uint8(self >> 8) }
func (self VidxRegisterH)  ID() uint8 { return uint8(self >> 8) }
func (self VidxRegister2H) ID() uint8 { return uint8(self >> 8) }
func (self VidxRegisterS)  ID() uint8 { return uint8(self >> 8) }
func (self VidxRegisterD)  ID() uint8 { return uint8(self >> 8) }

func (self VidxRegisterB)  Index() uint8 { return uint8(self & 0x1f) }
func (self VidxRegister4B) Index() uint8 { return uint8(self & 0x1f) }
func (self VidxRegisterH)  Index() uint8 { return uint8(self & 0x1f) }
func (self VidxRegister2H) Index() uint8 { return uint8(self & 0x1f) }
func (self VidxRegisterS)  Index() uint8 { return uint8(self & 0x1f) }
func (self VidxRegisterD)  Index() uint8 { return uint8(self & 0x1f) }

func (self VidxRegisterB)  String() string { return fmt.Sprintf("v%d.B[%d]" , self.ID(), self.Index()) }
func (self VidxRegister4B) String() string { return fmt.Sprintf("v%d.4B[%d]", self.ID(), self.Index()) }
func (self VidxRegisterH)  String() string { return fmt.Sprintf("v%d.H[%d]" , self.ID(), self.Index()) }
func (self VidxRegister2H) String() string { return fmt.Sprintf("v%d.2H[%d]", self.ID(), self.Index()) }
func (self VidxRegisterS)  String() string { return fmt.Sprintf("v%d.S[%d]" , self.ID(), self.Index()) }
func (self VidxRegisterD)  String() string { return fmt.Sprintf("v%d.D[%d]" , self.ID(), self.Index()) }
