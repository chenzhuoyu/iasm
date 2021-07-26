package obj

import (
    `os`
    `os/exec`
    `testing`

    `github.com/davecgh/go-spew/spew`
    `github.com/stretchr/testify/require`
)

const (
    _IMAGE_BASE = 0x00400000 // 64-bit virtual offsets always start at 0x00400000
)

func TestMachO_Create(t *testing.T) {
    fp, err := os.CreateTemp("", "macho_out-")
    require.NoError(t, err)
    code := []byte {
        0x48, 0xc7, 0xc7, 0x01, 0x00, 0x00, 0x00,   // MOVQ    $1, %rdi
        0x48, 0x8d, 0x35, 0x1b, 0x00, 0x00, 0x00,   // LEAQ    0x1b(%rip), %rsi
        0x48, 0xc7, 0xc2, 0x0e, 0x00, 0x00, 0x00,   // MOVQ    $14, %rdx
        0x48, 0xc7, 0xc0, 0x04, 0x00, 0x00, 0x02,   // MOVQ    $0x02000004, %rax
        0x0f, 0x05,                                 // SYSCALL
        0x31, 0xff,                                 // XORL    %edi, %edi
        0x48, 0xc7, 0xc0, 0x01, 0x00, 0x00, 0x02,   // MOVQ    $0x02000001, %rax
        0x0f, 0x05,                                 // SYSCALL
        'h', 'e', 'l', 'l', 'o', ',', ' ',
        'w', 'o', 'r', 'l', 'd', '\r', '\n',
    }
    err = assembleMachO(fp, code, _IMAGE_BASE, _IMAGE_BASE)
    require.NoError(t, err)
    err = fp.Close()
    require.NoError(t, err)
    err = os.Chmod(fp.Name(), 0755)
    require.NoError(t, err)
    println("Saved to", fp.Name())
    out, err := exec.Command(fp.Name()).Output()
    require.NoError(t, err)
    spew.Dump(out)
    require.Equal(t, []byte("hello, world\r\n"), out)
    err = os.Remove(fp.Name())
    require.NoError(t, err)
}
