package x86_64

import (
    `errors`
    `strconv`
    `unicode/utf8`
    `unsafe`

    `github.com/chenzhuoyu/iasm/asm`
)

const (
    _CC_digit = 1 << iota
    _CC_ident
    _CC_ident0
    _CC_number
)

func this(p *asm.Instruction) *Instruction {
    return (*Instruction)(unsafe.Pointer(p))
}

func brsize(p *asm.Instruction) int {
    switch p.Branch {
        case asm.BranchNone        : panic("p is not a branch")
        case asm.BranchAlways      : return _N_far_uncond
        case asm.BranchConditional : return _N_far_cond
        default                    : panic("invalid instruction")
    }
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
           (cc >= 'A' && cc <= 'F')
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
