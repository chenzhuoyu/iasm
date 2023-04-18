package asm

import (
    `encoding/binary`
    `errors`
    `reflect`
    `strconv`
    `unicode/utf8`
    `unsafe`

    `github.com/chenzhuoyu/iasm/internal/rt`
)

var (
    byteWrap = reflect.TypeOf(byte(0))
    byteType = (*rt.GoType)(rt.AsEface(byteWrap).Ptr)
)

//go:noescape
//go:linkname growslice runtime.growslice
func growslice(_ *rt.GoType, _ []byte, _ int) []byte

//go:noescape
//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
func memclrNoHeapPointers(_ unsafe.Pointer, _ uintptr)

func align(v int, n int) int {
    return (((v - 1) >> n) + 1) << n
}

func ispow2(v uint64) bool {
    return (v & (v - 1)) == 0
}

func isdigit(cc rune) bool {
    return '0' <= cc && cc <= '9'
}

func isalpha(cc rune) bool {
    return (cc >= 'a' && cc <= 'z') || (cc >= 'A' && cc <= 'Z')
}

func isident(cc rune) bool {
    return cc == '_' || isalpha(cc) || isdigit(cc)
}

func isident0(cc rune) bool {
    return cc == '_' || isalpha(cc)
}

func isnumber(cc rune) bool {
    return (cc == 'b' || cc == 'B') ||
           (cc == 'o' || cc == 'O') ||
           (cc == 'x' || cc == 'X') ||
           (cc >= '0' && cc <= '9') ||
           (cc >= 'a' && cc <= 'f') ||
           (cc >= 'A' && cc <= 'F') ||
           (cc == '.')
}

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

func literal64(v string) (uint64, error) {
    var nb int
    var ch rune
    var ex error
    var mm [12]byte

    /* unquote the runes */
    for v != "" {
        if ch, _, v, ex = strconv.UnquoteChar(v, '\''); ex != nil {
            return 0, ex
        } else if nb += utf8.EncodeRune(mm[nb:], ch); nb > 8 {
            return 0, errors.New("multi-char constant too large")
        }
    }

    /* convert to uint64 */
    return *(*uint64)(unsafe.Pointer(&mm)), nil
}

func expandmm(m *[]byte, n int, v byte) {
    sl := (*rt.GoSlice)(unsafe.Pointer(m))
    nb := sl.Len + n

    /* grow as needed */
    if nb > cap(*m) {
        *m = growslice(byteType, *m, nb)
    }

    /* fill the new area */
    memset(unsafe.Pointer(uintptr(sl.Ptr) + uintptr(sl.Len)), v, uintptr(n))
    sl.Len = nb
}

func memset(p unsafe.Pointer, c byte, n uintptr) {
    if c != 0 {
        memsetv(p, c, n)
    } else {
        memclrNoHeapPointers(p, n)
    }
}

func memsetv(p unsafe.Pointer, c byte, n uintptr) {
    for i := uintptr(0); i < n; i++ {
        *(*byte)(unsafe.Pointer(uintptr(p) + i)) = c
    }
}
