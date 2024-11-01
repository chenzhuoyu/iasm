package obj

import (
    `os`
    `os/exec`
    `testing`

    `github.com/davecgh/go-spew/spew`
    `github.com/stretchr/testify/require`
)

func TestOS_CompileAndLink(t *testing.T) {
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
    err := MacOS.CompileAndLink("/tmp/iasm-macho-out", x86_64, code, 0, 0)
    require.NoError(t, err)
    println("Saved to /tmp/iasm-macho-out")
    out, err := exec.Command("/tmp/iasm-macho-out").Output()
    require.NoError(t, err)
    spew.Dump(out)
    require.Equal(t, []byte("hello, world\r\n"), out)
    err = os.Remove("/tmp/iasm-macho-out")
    require.NoError(t, err)
}
