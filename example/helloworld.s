#define STDOUT      $1
#define IMAGE_BASE  0x04000000

#ifdef __Linux__
#define SYS_exit    $1
#define SYS_write   $4
#elif defined(__Darwin__)
#define SYS_exit    $0x02000001
#define SYS_write   $0x02000004
#else
#error Unsupported operating system.
#endif

.org   IMAGE_BASE
.entry start

start:
    movq    STDOUT, %rdi
    leaq    msg(%rip), %rsi
    movq    $13, %rdx
    movq    SYS_write, %rax
    syscall
    xorl    %edi, %edi
    movq    SYS_exit, %rax
    syscall

msg:
    .ascii "hello, world\n"
