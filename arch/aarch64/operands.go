package aarch64

// Indexing is the memory indexing modes.
type Indexing uint8

const (
    // NoIndexing represents the unsigned-offset form "[<Xn|SP>, #<pimm>]"
    NoIndexing Indexing = iota

    // PreIndexing represents the pre-indexing form "[<Xn|SP>, #<simm>]!"
    PreIndexing

    // PostIndexing represents the post-indexing form "[<Xn|SP>], #<simm>"
    PostIndexing
)

// Memory is the memory operand.
type Memory struct {
    Reg   Register64
    Imm   int32
    Index Indexing
}

func isReg    (x interface{}) bool { _, r := x.(Register)   ; return r }
func isReg32  (x interface{}) bool { _, r := x.(Register32) ; return r }
func isReg64  (x interface{}) bool { _, r := x.(Register64) ; return r }
func isWrOrSP (x interface{}) bool { x, r := x.(Register32) ; return r && x != WZR }
func isXrOrSP (x interface{}) bool { x, r := x.(Register64) ; return r && x != XZR }
