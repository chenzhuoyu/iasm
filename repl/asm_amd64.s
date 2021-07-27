#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

TEXT ·dumpregs(SB), (NOSPLIT | NOFRAME), $0
    NO_LOCAL_POINTERS
    PUSHFQ
    PUSHFQ
    PUSHQ AX
    MOVQ  0x20(SP), AX
    POPQ  (AX)
    MOVQ  BX, 0x08(AX)
    MOVQ  CX, 0x10(AX)
    MOVQ  DX, 0x18(AX)
    MOVQ  DI, 0x20(AX)
    MOVQ  SI, 0x28(AX)
    MOVQ  BP, 0x30(AX)
    MOVQ  SP, 0x38(AX)
    ADDQ  $16, 0x38(AX)
    MOVQ  R8, 0x40(AX)
    MOVQ  R9, 0x48(AX)
    MOVQ  R10, 0x50(AX)
    MOVQ  R11, 0x58(AX)
    MOVQ  R12, 0x60(AX)
    MOVQ  R13, 0x68(AX)
    MOVQ  R14, 0x70(AX)
    MOVQ  R15, 0x78(AX)
    MOVQ  16(SP), CX
    MOVQ  CX, 0x80(AX)
    POPQ  0x88(AX)
    MOVW  CS, 0x90(AX)
    MOVW  FS, 0x98(AX)
    MOVW  GS, 0xa0(AX)

    // SSE xmm0 ~ xmm15 registers
    MOVOU X0, 0xc0(AX)
    MOVOU X1, 0xd0(AX)
    MOVOU X2, 0xe0(AX)
    MOVOU X3, 0xf0(AX)
    MOVOU X4, 0x100(AX)
    MOVOU X5, 0x110(AX)
    MOVOU X6, 0x120(AX)
    MOVOU X7, 0x130(AX)
    MOVOU X8, 0x140(AX)
    MOVOU X9, 0x150(AX)
    MOVOU X10, 0x160(AX)
    MOVOU X11, 0x170(AX)
    MOVOU X12, 0x180(AX)
    MOVOU X13, 0x190(AX)
    MOVOU X14, 0x1a0(AX)
    MOVOU X15, 0x1b0(AX)

    // check for AVX
    CMPB ·hasAVX(SB), $0
    JE   _no_avx

    // AVX ymm0 ~ ymm15 registers
    VMOVDQU Y0, 0x2c0(AX)
    VMOVDQU Y1, 0x2e0(AX)
    VMOVDQU Y2, 0x300(AX)
    VMOVDQU Y3, 0x320(AX)
    VMOVDQU Y4, 0x340(AX)
    VMOVDQU Y5, 0x360(AX)
    VMOVDQU Y6, 0x380(AX)
    VMOVDQU Y7, 0x3a0(AX)
    VMOVDQU Y8, 0x3c0(AX)
    VMOVDQU Y9, 0x3e0(AX)
    VMOVDQU Y10, 0x400(AX)
    VMOVDQU Y11, 0x420(AX)
    VMOVDQU Y12, 0x440(AX)
    VMOVDQU Y13, 0x460(AX)
    VMOVDQU Y14, 0x480(AX)
    VMOVDQU Y15, 0x4a0(AX)

    // check for AVX512VL
    CMPB ·hasAVX512VL(SB), $0
    JE   _no_avx512vl

    // AVX512VL xmm16 ~ xmm31 registers
    LONG $0x087fe162; WORD $0x407f; BYTE $0x1c  // vmovdqu8 %xmm16, 0x1c0(%rax)
    LONG $0x087fe162; WORD $0x487f; BYTE $0x1d  // vmovdqu8 %xmm17, 0x1d0(%rax)
    LONG $0x087fe162; WORD $0x507f; BYTE $0x1e  // vmovdqu8 %xmm18, 0x1e0(%rax)
    LONG $0x087fe162; WORD $0x587f; BYTE $0x1f  // vmovdqu8 %xmm19, 0x1f0(%rax)
    LONG $0x087fe162; WORD $0x607f; BYTE $0x20  // vmovdqu8 %xmm20, 0x200(%rax)
    LONG $0x087fe162; WORD $0x687f; BYTE $0x21  // vmovdqu8 %xmm21, 0x210(%rax)
    LONG $0x087fe162; WORD $0x707f; BYTE $0x22  // vmovdqu8 %xmm22, 0x220(%rax)
    LONG $0x087fe162; WORD $0x787f; BYTE $0x23  // vmovdqu8 %xmm23, 0x230(%rax)
    LONG $0x087f6162; WORD $0x407f; BYTE $0x24  // vmovdqu8 %xmm24, 0x240(%rax)
    LONG $0x087f6162; WORD $0x487f; BYTE $0x25  // vmovdqu8 %xmm25, 0x250(%rax)
    LONG $0x087f6162; WORD $0x507f; BYTE $0x26  // vmovdqu8 %xmm26, 0x260(%rax)
    LONG $0x087f6162; WORD $0x587f; BYTE $0x27  // vmovdqu8 %xmm27, 0x270(%rax)
    LONG $0x087f6162; WORD $0x607f; BYTE $0x28  // vmovdqu8 %xmm28, 0x280(%rax)
    LONG $0x087f6162; WORD $0x687f; BYTE $0x29  // vmovdqu8 %xmm29, 0x290(%rax)
    LONG $0x087f6162; WORD $0x707f; BYTE $0x2a  // vmovdqu8 %xmm30, 0x2a0(%rax)
    LONG $0x087f6162; WORD $0x787f; BYTE $0x2b  // vmovdqu8 %xmm31, 0x2b0(%rax)

    // AVX512VL ymm16 ~ ymm31 registers
    LONG $0x28ffe162; WORD $0x407f; BYTE $0x26  // vmovdqu16 %ymm16, 0x4c0(%rax)
    LONG $0x28ffe162; WORD $0x487f; BYTE $0x27  // vmovdqu16 %ymm17, 0x4e0(%rax)
    LONG $0x28ffe162; WORD $0x507f; BYTE $0x28  // vmovdqu16 %ymm18, 0x500(%rax)
    LONG $0x28ffe162; WORD $0x587f; BYTE $0x29  // vmovdqu16 %ymm19, 0x520(%rax)
    LONG $0x28ffe162; WORD $0x607f; BYTE $0x2a  // vmovdqu16 %ymm20, 0x540(%rax)
    LONG $0x28ffe162; WORD $0x687f; BYTE $0x2b  // vmovdqu16 %ymm21, 0x560(%rax)
    LONG $0x28ffe162; WORD $0x707f; BYTE $0x2c  // vmovdqu16 %ymm22, 0x580(%rax)
    LONG $0x28ffe162; WORD $0x787f; BYTE $0x2d  // vmovdqu16 %ymm23, 0x5a0(%rax)
    LONG $0x28ff6162; WORD $0x407f; BYTE $0x2e  // vmovdqu16 %ymm24, 0x5c0(%rax)
    LONG $0x28ff6162; WORD $0x487f; BYTE $0x2f  // vmovdqu16 %ymm25, 0x5e0(%rax)
    LONG $0x28ff6162; WORD $0x507f; BYTE $0x30  // vmovdqu16 %ymm26, 0x600(%rax)
    LONG $0x28ff6162; WORD $0x587f; BYTE $0x31  // vmovdqu16 %ymm27, 0x620(%rax)
    LONG $0x28ff6162; WORD $0x607f; BYTE $0x32  // vmovdqu16 %ymm28, 0x640(%rax)
    LONG $0x28ff6162; WORD $0x687f; BYTE $0x33  // vmovdqu16 %ymm29, 0x660(%rax)
    LONG $0x28ff6162; WORD $0x707f; BYTE $0x34  // vmovdqu16 %ymm30, 0x680(%rax)
    LONG $0x28ff6162; WORD $0x787f; BYTE $0x35  // vmovdqu16 %ymm31, 0x6a0(%rax)

_no_avx512vl:
    CMPB ·hasAVX512F(SB), $0
    JE   _no_avx512f

    // AVX512F zmm0 ~ zmm31 registers
    LONG $0x487ef162; WORD $0x407f; BYTE $0x1b  // vmovdqu32 %zmm0, 0x6c0(%rax)
    LONG $0x487ef162; WORD $0x487f; BYTE $0x1c  // vmovdqu32 %zmm1, 0x700(%rax)
    LONG $0x487ef162; WORD $0x507f; BYTE $0x1d  // vmovdqu32 %zmm2, 0x740(%rax)
    LONG $0x487ef162; WORD $0x587f; BYTE $0x1e  // vmovdqu32 %zmm3, 0x780(%rax)
    LONG $0x487ef162; WORD $0x607f; BYTE $0x1f  // vmovdqu32 %zmm4, 0x7c0(%rax)
    LONG $0x487ef162; WORD $0x687f; BYTE $0x20  // vmovdqu32 %zmm5, 0x800(%rax)
    LONG $0x487ef162; WORD $0x707f; BYTE $0x21  // vmovdqu32 %zmm6, 0x840(%rax)
    LONG $0x487ef162; WORD $0x787f; BYTE $0x22  // vmovdqu32 %zmm7, 0x880(%rax)
    LONG $0x487e7162; WORD $0x407f; BYTE $0x23  // vmovdqu32 %zmm8, 0x8c0(%rax)
    LONG $0x487e7162; WORD $0x487f; BYTE $0x24  // vmovdqu32 %zmm9, 0x900(%rax)
    LONG $0x487e7162; WORD $0x507f; BYTE $0x25  // vmovdqu32 %zmm10, 0x940(%rax)
    LONG $0x487e7162; WORD $0x587f; BYTE $0x26  // vmovdqu32 %zmm11, 0x980(%rax)
    LONG $0x487e7162; WORD $0x607f; BYTE $0x27  // vmovdqu32 %zmm12, 0x9c0(%rax)
    LONG $0x487e7162; WORD $0x687f; BYTE $0x28  // vmovdqu32 %zmm13, 0xa00(%rax)
    LONG $0x487e7162; WORD $0x707f; BYTE $0x29  // vmovdqu32 %zmm14, 0xa40(%rax)
    LONG $0x487e7162; WORD $0x787f; BYTE $0x2a  // vmovdqu32 %zmm15, 0xa80(%rax)
    LONG $0x487ee162; WORD $0x407f; BYTE $0x2b  // vmovdqu32 %zmm16, 0xac0(%rax)
    LONG $0x487ee162; WORD $0x487f; BYTE $0x2c  // vmovdqu32 %zmm17, 0xb00(%rax)
    LONG $0x487ee162; WORD $0x507f; BYTE $0x2d  // vmovdqu32 %zmm18, 0xb40(%rax)
    LONG $0x487ee162; WORD $0x587f; BYTE $0x2e  // vmovdqu32 %zmm19, 0xb80(%rax)
    LONG $0x487ee162; WORD $0x607f; BYTE $0x2f  // vmovdqu32 %zmm20, 0xbc0(%rax)
    LONG $0x487ee162; WORD $0x687f; BYTE $0x30  // vmovdqu32 %zmm21, 0xc00(%rax)
    LONG $0x487ee162; WORD $0x707f; BYTE $0x31  // vmovdqu32 %zmm22, 0xc40(%rax)
    LONG $0x487ee162; WORD $0x787f; BYTE $0x32  // vmovdqu32 %zmm23, 0xc80(%rax)
    LONG $0x487e6162; WORD $0x407f; BYTE $0x33  // vmovdqu32 %zmm24, 0xcc0(%rax)
    LONG $0x487e6162; WORD $0x487f; BYTE $0x34  // vmovdqu32 %zmm25, 0xd00(%rax)
    LONG $0x487e6162; WORD $0x507f; BYTE $0x35  // vmovdqu32 %zmm26, 0xd40(%rax)
    LONG $0x487e6162; WORD $0x587f; BYTE $0x36  // vmovdqu32 %zmm27, 0xd80(%rax)
    LONG $0x487e6162; WORD $0x607f; BYTE $0x37  // vmovdqu32 %zmm28, 0xdc0(%rax)
    LONG $0x487e6162; WORD $0x687f; BYTE $0x38  // vmovdqu32 %zmm29, 0xe00(%rax)
    LONG $0x487e6162; WORD $0x707f; BYTE $0x39  // vmovdqu32 %zmm30, 0xe40(%rax)
    LONG $0x487e6162; WORD $0x787f; BYTE $0x3a  // vmovdqu32 %zmm31, 0xe80(%rax)

    // check for AVX512BW
    CMPB ·hasAVX512BW(SB), $0
    JE   _no_avx512bw

    // AVX512BW 64-bit K registers
    QUAD $0x000ec08091f8e1c4; BYTE $0x00    // kmovq %k0, 0xec0(%rax)
    QUAD $0x000ec88891f8e1c4; BYTE $0x00    // kmovq %k1, 0xec8(%rax)
    QUAD $0x000ed09091f8e1c4; BYTE $0x00    // kmovq %k2, 0xed0(%rax)
    QUAD $0x000ed89891f8e1c4; BYTE $0x00    // kmovq %k3, 0xed8(%rax)
    QUAD $0x000ee0a091f8e1c4; BYTE $0x00    // kmovq %k4, 0xee0(%rax)
    QUAD $0x000ee8a891f8e1c4; BYTE $0x00    // kmovq %k5, 0xee8(%rax)
    QUAD $0x000ef0b091f8e1c4; BYTE $0x00    // kmovq %k6, 0xef0(%rax)
    QUAD $0x000ef8b891f8e1c4; BYTE $0x00    // kmovq %k7, 0xef8(%rax)
    JMP  _avx512bw_done

_no_avx512bw:
    QUAD $0x00000ec08091f8c5    // kmovw %k0, 0xec0(%rax)
    QUAD $0x00000ec88891f8c5    // kmovw %k1, 0xec8(%rax)
    QUAD $0x00000ed09091f8c5    // kmovw %k2, 0xed0(%rax)
    QUAD $0x00000ed89891f8c5    // kmovw %k3, 0xed8(%rax)
    QUAD $0x00000ee0a091f8c5    // kmovw %k4, 0xee0(%rax)
    QUAD $0x00000ee8a891f8c5    // kmovw %k5, 0xee8(%rax)
    QUAD $0x00000ef0b091f8c5    // kmovw %k6, 0xef0(%rax)
    QUAD $0x00000ef8b891f8c5    // kmovw %k7, 0xef8(%rax)

_avx512bw_done:
_no_avx512f:
_no_avx:
    MOVQ 0x10(AX), CX
    MOVQ (AX), AX
    POPFQ
    RET

TEXT ·execaddr(SB), (NOSPLIT | NOFRAME), $0
    NO_LOCAL_POINTERS
    LONG  $0x102474ff       // pushq 0x10(%rsp)
    CALL  ·dumpregs(SB)
    LEAQ  8(SP), SP
    CALL  exectrampoline(SB)
    LONG  $0x182474ff       // pushq 0x18(%rsp)
    CALL  ·dumpregs(SB)
    LEAQ  8(SP), SP
    RET

TEXT exectrampoline(SB), (NOSPLIT | NOFRAME), $0
    NO_LOCAL_POINTERS
    JMP 0x10(SP)
