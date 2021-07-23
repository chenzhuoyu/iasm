package expr

var op1ch = [...]bool {
    '+': true,
    '-': true,
    '*': true,
    '/': true,
    '%': true,
    '&': true,
    '|': true,
    '^': true,
    '~': true,
    '(': true,
    ')': true,
}

var op2ch = [...]bool {
    '*': true,
    '<': true,
    '>': true,
}

func neg2(v int64, err error) (int64, error) {
    if err != nil {
        return 0, err
    } else {
        return -v, nil
    }
}

func inv2(v int64, err error) (int64, error) {
    if err != nil {
        return 0, err
    } else {
        return ^v, nil
    }
}

func isop1ch(ch rune) bool {
    return ch >= 0 && int(ch) < len(op1ch) && op1ch[ch]
}

func isop2ch(ch rune) bool {
    return ch >= 0 && int(ch) < len(op2ch) && op2ch[ch]
}

func isdigit(ch rune) bool {
    return ch >= '0' && ch <= '9'
}

func isident(ch rune) bool {
    return isdigit(ch) || isident0(ch)
}

func isident0(ch rune) bool {
    return (ch == '_') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func ishexdigit(ch rune) bool {
    return isdigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}
