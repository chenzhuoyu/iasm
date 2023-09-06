package rt

import (
    `reflect`
    `unsafe`
)

type GoType struct {
    Size       uintptr
    PtrData    uintptr
    Hash       uint32
    Flags      uint8
    Align      uint8
    FieldAlign uint8
    KindFlags  uint8
    Traits     unsafe.Pointer
    GCData     *byte
    Str        int32
    PtrTy      int32
}

const (
    _KindMask = (1 << 5) - 1
)

// This lookup table must be large enough to hold all the `reflect.Kind` values.
var _SignedTy = [27]bool {
    reflect.Int   : true,
    reflect.Int8  : true,
    reflect.Int16 : true,
    reflect.Int32 : true,
    reflect.Int64 : true,
}

func (self *GoType) Kind() reflect.Kind {
    return reflect.Kind(self.KindFlags & _KindMask)
}

type GoSlice struct {
    Ptr unsafe.Pointer
    Len int
    Cap int
}

type GoEface struct {
    Ty  *GoType
    Ptr unsafe.Pointer
}

func (self GoEface) s64(ty string) int64 {
    switch self.Ty.Size {
        case 1  : return int64(*(*int8)(self.Ptr))
        case 2  : return int64(*(*int16)(self.Ptr))
        case 4  : return int64(*(*int32)(self.Ptr))
        case 8  : return *(*int64)(self.Ptr)
        default : panic("cannot convert " + self.Kind().String() + " to " + ty)
    }
}

func (self GoEface) u64(ty string) uint64 {
    switch self.Ty.Size {
        case 1  : return uint64(*(*uint8)(self.Ptr))
        case 2  : return uint64(*(*uint16)(self.Ptr))
        case 4  : return uint64(*(*uint32)(self.Ptr))
        case 8  : return *(*uint64)(self.Ptr)
        default : panic("cannot convert " + self.Kind().String() + " to " + ty)
    }
}

func (self GoEface) Kind() reflect.Kind {
    if self.Ty != nil {
        return self.Ty.Kind()
    } else {
        return reflect.Invalid
    }
}

func (self GoEface) Pack() (ret interface{}) {
    *(*GoEface)(unsafe.Pointer(&ret)) = self
    return
}

func (self GoEface) ToInt64() int64 {
    if _SignedTy[self.Ty.Kind()] {
        return self.s64("int64")
    } else {
        return int64(self.u64("int64"))
    }
}

func (self GoEface) ToUint64() uint64 {
    if _SignedTy[self.Ty.Kind()] {
        return uint64(self.s64("uint64"))
    } else {
        return self.u64("uint64")
    }
}

func (self GoEface) ToFloat64() float64 {
    switch self.Ty.Size {
        case 4  : return float64(*(*float32)(self.Ptr))
        case 8  : return *(*float64)(self.Ptr)
        default : panic("cannot convert " + self.Kind().String() + " to float64")
    }
}

func TypeOf(v interface{}) *GoType {
    return AsEface(v).Ty
}

func AsEface(v interface{}) GoEface {
    return *(*GoEface)(unsafe.Pointer(&v))
}
