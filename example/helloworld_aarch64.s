#define STDOUT      #1
#define IMAGE_BASE  0x04000000

#ifdef __Linux__
#define SVC_ID      #0
#define SVC_REG     x8
#define SYS_exit    #93
#define SYS_write   #64
#elif defined(__Darwin__)
#define SVC_ID      #0x80
#define SVC_REG     x16
#define SYS_exit    #1
#define SYS_write   #4
#else
#error Unsupported operating system.
#endif

.org   IMAGE_BASE
.entry start

start:
    mov     x0, #1
    ldr     x1, =msg
    mov     x2, #13
    mov     SVC_REG, SYS_write
    svc     SVC_ID
    mov     x0, #0
    mov     SVC_REG, SYS_exit
    svc     SVC_ID

msg:
    .ascii "hello, world\n"
