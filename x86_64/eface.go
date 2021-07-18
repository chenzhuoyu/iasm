package x86_64

import (
    `reflect`
    `unsafe`
)

type _GoType struct {
    size   uintptr
    pdata  uintptr
    hash   uint32
    flags  uint8
    align  uint8
    falign uint8
    kflags uint8
    traits unsafe.Pointer
    gcdata *byte
    str    int32
    ptrx   int32
}

const (
    _KindMask = (1 << 5) - 1
)

func (self *_GoType) kind() reflect.Kind {
    return reflect.Kind(self.kflags & _KindMask)
}

type _GoEface struct {
    vt  *_GoType
    ptr unsafe.Pointer
}

func (self *_GoEface) kind() reflect.Kind {
    if self.vt != nil {
        return self.vt.kind()
    } else {
        return reflect.Invalid
    }
}

func (self *_GoEface) toInt64() int64 {
    if self.vt.size == 8 {
        return *(*int64)(self.ptr)
    } else if self.vt.size == 4 {
        return int64(*(*int32)(self.ptr))
    } else if self.vt.size == 2 {
        return int64(*(*int16)(self.ptr))
    } else {
        return int64(*(*int8)(self.ptr))
    }
}

func efaceOf(v interface{}) _GoEface {
    return *(*_GoEface)(unsafe.Pointer(&v))
}
