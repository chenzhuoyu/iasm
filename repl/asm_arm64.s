#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

TEXT ·dumpregs(SB), (NOSPLIT|NOFRAME), $0
    NO_LOCAL_POINTERS
    SUB     $32, RSP, RSP
    STP     (R29, R30), -8(RSP)
    SUB     $8, RSP, R29
    MOVD    R0, 8(RSP)
    MOVD    40(RSP), R0
    STP     (R0, R1), 0x00(R0)
    STP     (R2, R3), 0x10(R0)
    STP     (R4, R5), 0x20(R0)
    STP     (R6, R7), 0x30(R0)
    STP     (R8, R9), 0x40(R0)
    STP     (R10, R11), 0x50(R0)
    STP     (R12, R13), 0x60(R0)
    STP     (R14, R15), 0x70(R0)
    STP     (R16, R17), 0x80(R0)
    WORD    $0xa9094c12         // stp x18, x19, [x0, #0x90]
    STP     (R20, R21), 0xa0(R0)
    STP     (R22, R23), 0xb0(R0)
    STP     (R24, R25), 0xc0(R0)
    STP     (R26, R27), 0xd0(R0)
    WORD    $0xa90e741c         // stp x28, x29, [x0, #0xe0]
    MOVD    8(RSP), R16
    MOVD    R16, (R0)
    MOVD    RSP, R16
    ADD     $0xf0, R0, R0
    STP.P   (R30, R16), 0x10(R0)
    VST1.P  [V0.B16, V1.B16, V2.B16, V3.B16], 0x40(R0)
    VST1.P  [V4.B16, V5.B16, V6.B16, V7.B16], 0x40(R0)
    VST1.P  [V8.B16, V9.B16, V10.B16, V11.B16], 0x40(R0)
    VST1.P  [V12.B16, V13.B16, V14.B16, V15.B16], 0x40(R0)
    VST1.P  [V16.B16, V17.B16, V18.B16, V19.B16], 0x40(R0)
    VST1.P  [V20.B16, V21.B16, V22.B16, V23.B16], 0x40(R0)
    VST1.P  [V24.B16, V25.B16, V26.B16, V27.B16], 0x40(R0)
    VST1.P  [V28.B16, V29.B16, V30.B16, V31.B16], 0x40(R0)
    ADR     0(PC), R16
    MRS     NZCV, R17
    STP     (R16, R17), (R0)
    LDP     -0x280(R0), (R16, R17)
    MOVD    8(RSP), R0
    LDP     -8(RSP), (R29, R30)
    ADD     $32, RSP, RSP
    RET

TEXT ·execaddr(SB), (NOSPLIT|NOFRAME), $0
    NO_LOCAL_POINTERS
    SUB     $32, RSP, RSP
    STP     (R29, R30), -8(RSP)
    SUB     $8, RSP, R29
    MOVD    R0, 16(RSP)
    MOVD    48(RSP), R0
    MOVD    R0, 8(RSP)
    MOVD    16(RSP), R0
    BL      ·dumpregs(SB)
    MOVD    40(RSP), R0
    CALL    R0
    MOVD    R0, 16(RSP)
    MOVD    56(RSP), R0
    MOVD    R0, 8(RSP)
    MOVD    16(RSP), R0
    BL      ·dumpregs(SB)
    LDP     -8(RSP), (R29, R30)
    ADD     $32, RSP, RSP
    RET
