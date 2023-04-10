package asm

import (
    `errors`
    `fmt`
    `sync/atomic`

    `github.com/chenzhuoyu/iasm/expr`
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Register represents a hardware register.
type Register interface {
    tag.Sealed
    fmt.Stringer
    ID() uint8
}

// Addressable is a union to represent an addressable operand.
type Addressable interface {
    tag.Sealed
    fmt.Stringer
    tag.Disposable
    tag.Verifiable
    Addressable()
}

// Label represents a location within the program.
type Label struct {
    refs int64
    Name string
    Dest *Instruction
}

func (*Label) Sealed(tag.Tag) {}
func (*Label) EnsureValid()   {}
func (*Label) Addressable()   {}

func (self *Label) drop() {
    if self.Dest != nil {
        self.Dest.Free()
    }
}

func (self *Label) Free() {
    if atomic.AddInt64(&self.refs, -1) == 0 {
        self.drop()
        freeLabel(self)
    }
}

func (self *Label) String() string {
    if self.Dest == nil {
        return self.Name
    } else {
        return fmt.Sprintf("%s@%#x", self.Name, self.Dest.PC)
    }
}

func (self *Label) Retain() *Label {
    atomic.AddInt64(&self.refs, 1)
    return self
}

func (self *Label) Evaluate() (int64, error) {
    if self.Dest != nil {
        return int64(self.Dest.PC), nil
    } else {
        return 0, errors.New("unresolved label: " + self.Name)
    }
}

func (self *Label) RelativeTo(pc uintptr) uintptr {
    if self.Dest == nil {
        panic("aarch64: unresolved label: " + self.Name)
    } else {
        return self.Dest.PC - pc
    }
}

// MemoryAddressExtension represents an arch-specific memory address extension.
type MemoryAddressExtension interface {
    tag.Sealed
    tag.Disposable
    String(MemoryAddress) string
    EnsureValid(MemoryAddress)
    MemoryAddressExtension()
}

// MemoryAddress represents a memory address.
type MemoryAddress struct {
    Base   Register
    Index  Register
    Offset int32
    Ext    MemoryAddressExtension
}

func (MemoryAddress) Sealed(tag.Tag) {}
func (MemoryAddress) Addressable()   {}

func (self MemoryAddress) Free() {
    self.Ext.Free()
}

func (self MemoryAddress) String() string {
    if self.Ext == nil {
        return "(invalid)"
    } else {
        return self.Ext.String(self)
    }
}

func (self MemoryAddress) EnsureValid() {
    self.Ext.EnsureValid(self)
}

// RelativeOffset represents an PC-relative offset.
type RelativeOffset int32

func ComputeOffset(term expr.Term, pc uintptr, n int) RelativeOffset {
    if val, err := term.Evaluate(); err != nil {
        panic("cannot evaluate: " + err.Error())
    } else {
        return RelativeOffset(uintptr(val) - pc - uintptr(n))
    }
}

func (RelativeOffset) Sealed(tag.Tag) {}
func (RelativeOffset) Free()          {}
func (RelativeOffset) EnsureValid()   {}
func (RelativeOffset) Addressable()   {}

func (self RelativeOffset) String() string {
    if self == 0 {
        return "."
    } else if self > 0 {
        return fmt.Sprintf(".+%d", self)
    } else {
        return fmt.Sprintf(".-%d", -self)
    }
}

// MemoryOperandExtension represents an arch-specific memory operand extension.
type MemoryOperandExtension interface {
    tag.Sealed
    tag.Disposable
    String(*MemoryOperand) string
    EnsureValid(*MemoryOperand)
    MemoryOperandExtension()
}

// MemoryOperand represents a memory operand for an instruction.
type MemoryOperand struct {
    refs int64
    Addr Addressable
    Ext  MemoryOperandExtension
}

func (self *MemoryOperand) Free() {
    if atomic.AddInt64(&self.refs, -1) == 0 {
        self.Ext.Free()
        self.Addr.Free()
        freeMemoryOperand(self)
    }
}

func (self *MemoryOperand) String() string {
    if self.Ext == nil {
        return "(invalid)"
    } else {
        return self.Ext.String(self)
    }
}

func (self *MemoryOperand) Retain() *MemoryOperand {
    atomic.AddInt64(&self.refs, 1)
    return self
}

func (self *MemoryOperand) EnsureValid() {
    self.Ext.EnsureValid(self)
}
