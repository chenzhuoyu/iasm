package x86_64

import (
    `encoding/binary`
)

func append8(m *[]byte, v byte) {
    *m = append(*m, v)
}

func append16(m *[]byte, v uint16) {
    p := len(*m)
    *m = append(*m, 0, 0)
    binary.LittleEndian.PutUint16((*m)[p:], v)
}

func append32(m *[]byte, v uint32) {
    p := len(*m)
    *m = append(*m, 0, 0, 0, 0)
    binary.LittleEndian.PutUint32((*m)[p:], v)
}

func append64(m *[]byte, v uint64) {
    p := len(*m)
    *m = append(*m, 0, 0, 0, 0, 0, 0, 0, 0)
    binary.LittleEndian.PutUint64((*m)[p:], v)
}
