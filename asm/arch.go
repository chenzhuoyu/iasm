package asm

import (
    `fmt`
    `math`
    `sync`

    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Feature represents one of the optional CPU features.
type Feature interface {
    tag.Sealed
    fmt.Stringer
    FeatureID() uint8
}

// Implementation is the arch-specific implementations.
type Implementation interface {
    tag.Sealed
    New() *Instruction
    Free(p *Instruction)
    Encode(p *Instruction, m *[]byte) int
    Assemble(p *Program, pc uintptr) []byte
}

// Arch represents an CPU architecture.
type Arch struct {
    fmax uint8
    fset [4]uint64
    impl Implementation
}

var (
    archmut = new(sync.RWMutex)
    archtab = make(map[string]*Arch)
)

var featmax = [4]uint64 {
    math.MaxUint64,
    math.MaxUint64,
    math.MaxUint64,
    math.MaxUint64,
}

// GetArch finds the architecture by name.
func GetArch(name string) (p *Arch) {
    archmut.RLock()
    p = archtab[name]
    archmut.RUnlock()
    return
}

// RegisterArch adds a new Arch into the registry with all features enabled.
func RegisterArch(name string, maxft Feature, impl Implementation) (p *Arch) {
    p = new(Arch)
    p.impl = impl
    p.fset = featmax
    p.fmax = maxft.FeatureID()
    archmut.Lock()
    archtab[name] = p
    archmut.Unlock()
    return
}

// Enable marks the feature as enabled.
func (self *Arch) Enable(ft Feature) *Arch {
    if id := ft.FeatureID(); id > self.fmax {
        panic("iasm: no such feature: " + ft.String())
    } else {
        self.fset[id >> 6] |= 1 << (id & 0x3f)
        return self
    }
}

// Disable marks the feature as disabled.
func (self *Arch) Disable(ft Feature) *Arch {
    if id := ft.FeatureID(); id > self.fmax {
        panic("iasm: no such feature: " + ft.String())
    } else {
        self.fset[id >> 6] &^= 1 << (id & 0x3f)
        return self
    }
}

// Require ensures that the specified feature is enabled, panic if not.
func (self *Arch) Require(ft Feature) {
    if !self.HasFeature(ft) {
        panic("Feature '" + ft.String() + "' was not enabled")
    }
}

// HasFeature checks if certain feature is valid and enabled.
func (self *Arch) HasFeature(ft Feature) bool {
    id := ft.FeatureID()
    return id <= self.fmax && (self.fset[id >> 6] & (1 << (id & 0x3f))) != 0
}

// CreateProgram creates a new arch-specific Program.
func (self *Arch) CreateProgram() *Program {
    return newProgram(self)
}
