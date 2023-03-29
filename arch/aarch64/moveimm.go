package aarch64

func encodeMoveImm(v interface{}, bits uint64, inverted bool) (uint32, bool) {
    var ok  bool
    var nb  uint64
    var imm uint64
    var lsb uint64
    var msb uint64

    /* sanity check */
    if bits != 32 && bits != 64 {
        panic("aarch64: moveimm: invalid bits")
    }

    /* convert to uint64 */
    if imm, ok = asUint64(v); !ok {
        return 0, false
    }

    /* encode 0 specifially as move wide immediate (non-inverted) */
    if inverted && imm == 0 {
        return 0, false
    }

    /* excluding 0xffff0000 and 0x0000ffff for move 32-bit inverted wide immediate */
    if inverted && bits == 32 && (imm == 0xffff0000 || imm == 0x0000ffff) {
        return 0, false
    }

    /* invert as needed, mask the high bits if it's a 32-bit number */
    if inverted {
        if imm = ^imm; bits == 32 {
            imm &= 0xffffffff
        }
    }

    /* scan with 16-bit step */
    for nb, ok = 0, false; nb < bits; nb += 16 {
        msb = imm >> nb
        lsb = imm & ((1 << nb) - 1)

        /* check if it can be encoded */
        if lsb == 0 && (msb &^ 0xffff) == 0 {
            ok = true
            break
        }
    }

    /* encode the integer if possible */
    if !ok {
        return 0, false
    } else {
        return uint32((nb << 12) | msb), true
    }
}
