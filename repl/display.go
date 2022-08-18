package repl

import (
    `fmt`
    `strings`
    `unicode`
)

func vector(v []byte, indent int) string {
    n := len(v)
    m := make([]string, 0, len(v))

    /* convert each element */
    for i := 0; i < n; i++ {
        if i == 0 || (i & 15) != 0 {
            m = append(m, fmt.Sprintf("0x%02x", v[n - i - 1]))
        } else {
            m = append(m, fmt.Sprintf("\n%s0x%02x", strings.Repeat(" ", indent + 1), v[n - i - 1]))
        }
    }

    /* join them together */
    return fmt.Sprintf(
        "{%s}",
        strings.Join(m, ", "),
    )
}

func display(v byte) string {
    if !unicode.IsPrint(rune(v)) {
        return "."
    } else {
        return string(v)
    }
}

func vecdiff(v0 []byte, v1 []byte, indent int) string {
    return vector(v0, indent) + strings.Repeat(" ", indent - 2) + "->" + vector(v1, indent)
}

func asmdump(m []byte, pos uintptr, src string) string {
    row := -1
    ret := []string(nil)

    /* must have at least 1 byte */
    if len(m) == 0 {
        panic("empty instruction bytes")
    }

    /* convert all the bytes */
    for i, v := range m {
        if i % 7 != 0 {
            ret[row] = ret[row] + fmt.Sprintf(" %02x", v)
        } else {
            ret, row = append(ret, fmt.Sprintf("(%#06x) %02x", pos + uintptr(i), v)), row + 1
        }
    }

    /* pad the first line if needed */
    if n := len(m); n < 7 {
        ret[0] += strings.Repeat(" ", (7 - n) * 3)
    }

    /* add the source */
    ret[0] += "    " + src
    return strings.Join(ret, "\n")
}

func hexdump(buf []byte, start uintptr) string {
    off := 0
    nbs := len(buf)
    ret := []string(nil)

    /* dump the data, 16-byte loop */
    for nbs >= 16 {
        _ = buf[off + 15]
        ret = append(ret, fmt.Sprintf("%x:", start + uintptr(off)))
        ret = append(ret, fmt.Sprintf(" %02x %02x %02x %02x", buf[off +  0], buf[off +  1], buf[off +  2], buf[off +  3]))
        ret = append(ret, fmt.Sprintf(" %02x %02x %02x %02x", buf[off +  4], buf[off +  5], buf[off +  6], buf[off +  7]), " ")
        ret = append(ret, fmt.Sprintf(" %02x %02x %02x %02x", buf[off +  8], buf[off +  9], buf[off + 10], buf[off + 11]))
        ret = append(ret, fmt.Sprintf(" %02x %02x %02x %02x", buf[off + 12], buf[off + 13], buf[off + 14], buf[off + 15]), "  ")

        /* print the 16 characters */
        for i := 0; i < 16; i++ {
            ret = append(ret, display(buf[off + i]))
        }

        /* move to the next line */
        off = off + 16
        nbs = nbs - 16
        ret = append(ret, "\n")
    }

    /* still have more bytes */
    if nbs > 0 {
        ret = append(ret, fmt.Sprintf("%x:", start + uintptr(off)))
    }

    /* print the remaining bytes */
    for i := 0; i < nbs; i++ {
        if i != 7 {
            ret = append(ret, fmt.Sprintf(" %02x", buf[off + i]))
        } else {
            ret = append(ret, fmt.Sprintf(" %02x ", buf[off + i]))
        }
    }

    /* pad the last line */
    if nbs > 0 {
        if nbs < 8 {
            ret = append(ret, strings.Repeat(" ", (17 - int(nbs)) * 3))
        } else {
            ret = append(ret, strings.Repeat(" ", (17 - int(nbs)) * 3 - 1))
        }
    }

    /* print the last characters */
    for i := 0; i < nbs; i++ {
        ret = append(ret, display(buf[off + i]))
    }

    /* check the final terminator */
    if nbs <= 0 {
        return strings.Join(ret, "")
    } else {
        return strings.Join(append(ret, "\n"), "")
    }
}
