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

func (self *GoEface) Kind() reflect.Kind {
    if self.Ty != nil {
        return self.Ty.Kind()
    } else {
        return reflect.Invalid
    }
}

func (self *GoEface) ToInt64() int64 {
    if self.Ty.Size == 8 {
        return *(*int64)(self.Ptr)
    } else if self.Ty.Size == 4 {
        return int64(*(*int32)(self.Ptr))
    } else if self.Ty.Size == 2 {
        return int64(*(*int16)(self.Ptr))
    } else {
        return int64(*(*int8)(self.Ptr))
    }
}

func AsEface(v interface{}) GoEface {
    return *(*GoEface)(unsafe.Pointer(&v))
}
