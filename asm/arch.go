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

// ProgramFactory is the factory of an arch-specific Program.
type ProgramFactory interface {
    tag.Sealed
    CreateProgram() Program
}

// Arch represents an CPU architecture.
type Arch struct {
    fmax uint8
    fset [4]uint64
    pfac ProgramFactory
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
func RegisterArch(name string, maxft Feature, pfac ProgramFactory) (p *Arch) {
    p = new(Arch)
    p.pfac = pfac
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

// HasFeature checks if certain feature is valid and enabled.
func (self *Arch) HasFeature(ft Feature) bool {
    id := ft.FeatureID()
    return id <= self.fmax && (self.fset[id >> 6] & (1 << (id & 0x3f))) != 0
}

// CreateProgram creates a new arch-specific Program.
func (self *Arch) CreateProgram() Program {
    return self.pfac.CreateProgram()
}
