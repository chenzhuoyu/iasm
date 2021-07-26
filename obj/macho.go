package obj

import (
    `encoding/binary`
    `io`
    `unsafe`
)

type MachOHeader struct {
    Magic      uint32
    CPUType    uint32
    CPUSubType uint32
    FileType   uint32
    CmdCount   uint32
    CmdSize    uint32
    Flags      uint32
    _          uint32
}

type SegmentCommand struct {
    Cmd          uint32
    Size         uint32
    Name         [16]byte
    VMAddr       uint64
    VMSize       uint64
    FileOffset   uint64
    FileSize     uint64
    MaxProtect   uint32
    InitProtect  uint32
    SectionCount uint32
    Flags        uint32
}

type Registers struct {
    RAX    uint64
    RBX    uint64
    RCX    uint64
    RDX    uint64
    RDI    uint64
    RSI    uint64
    RBP    uint64
    RSP    uint64
    R8     uint64
    R9     uint64
    R10    uint64
    R11    uint64
    R12    uint64
    R13    uint64
    R14    uint64
    R15    uint64
    RIP    uint64
    RFLAGS uint64
    CS     uint64
    FS     uint64
    GS     uint64
}

type UnixThreadCommand struct {
    Cmd    uint32
    Size   uint32
    Flavor uint32
    Count  uint32
    Regs   Registers
}

const (
    _MH_MAGIC_64 = 0xfeedfacf
    _MH_EXECUTE  = 0x02
    _MH_NOUNDEFS = 0x01
)

const (
    _CPU_TYPE_I386  = 0x00000007
    _CPU_ARCH_ABI64 = 0x01000000
)

const (
    _CPU_SUBTYPE_LIB64    = 0x80000000
    _CPU_SUBTYPE_I386_ALL = 0x00000003
)

const (
    _LC_SEGMENT_64 = 0x19
    _LC_UNIXTHREAD = 0x05
)

const (
    _VM_PROT_READ    = 0x01
    _VM_PROT_WRITE   = 0x02
    _VM_PROT_EXECUTE = 0x04
)

const (
    _x86_THREAD_STATE64          = 0x04
    _x86_EXCEPTION_STATE64_COUNT = 42
)

const (
    _MACHO_SIZE      = uint32(unsafe.Sizeof(MachOHeader{}))
    _SEGMENT_SIZE    = uint32(unsafe.Sizeof(SegmentCommand{}))
    _UNIXTHREAD_SIZE = uint32(unsafe.Sizeof(UnixThreadCommand{}))
)

var (
    zeroBytes = [4096]byte{}
)

func assembleMachO(w io.Writer, code []byte, base uint64, entry uint64) error {
    var p0name [16]byte
    var txname [16]byte

    /* segment names */
    copy(txname[:], "__TEXT")
    copy(p0name[:], "__PAGEZERO")

    /* calculate size of code */
    clen := uint64(len(code))
    hlen := uint64(_MACHO_SIZE) + uint64(_SEGMENT_SIZE) + uint64(_SEGMENT_SIZE) + uint64(_UNIXTHREAD_SIZE)

    /* Page-0 Segment */
    p0 := SegmentCommand {
        Cmd    : _LC_SEGMENT_64,
        Size   : _SEGMENT_SIZE,
        Name   : p0name,
        VMSize : base,
    }

    /* TEXT Segment */
    text := SegmentCommand {
        Cmd         : _LC_SEGMENT_64,
        Size        : _SEGMENT_SIZE,
        Name        : txname,
        VMAddr      : base,
        VMSize      : hlen + clen,
        FileSize    : hlen + clen,
        MaxProtect  : _VM_PROT_READ | _VM_PROT_WRITE | _VM_PROT_EXECUTE,
        InitProtect : _VM_PROT_READ | _VM_PROT_EXECUTE,
    }

    /* UNIX Thread Metadata */
    unix := UnixThreadCommand {
        Cmd    : _LC_UNIXTHREAD,
        Size   : _UNIXTHREAD_SIZE,
        Flavor : _x86_THREAD_STATE64,
        Count  : _x86_EXCEPTION_STATE64_COUNT,
        Regs   : Registers{RIP: hlen + entry},
    }

    /* Mach-O Header */
    macho := MachOHeader {
        Magic      : _MH_MAGIC_64,
        CPUType    : _CPU_ARCH_ABI64 | _CPU_TYPE_I386,
        CPUSubType : _CPU_SUBTYPE_LIB64 | _CPU_SUBTYPE_I386_ALL,
        FileType   : _MH_EXECUTE,
        CmdCount   : 3,
        CmdSize    : _SEGMENT_SIZE * 2 + _UNIXTHREAD_SIZE,
        Flags      : _MH_NOUNDEFS,
    }

    /* write the headers */
    if err := binary.Write(w, binary.LittleEndian, &macho) ; err != nil { return err }
    if err := binary.Write(w, binary.LittleEndian, &p0)    ; err != nil { return err }
    if err := binary.Write(w, binary.LittleEndian, &text)  ; err != nil { return err }
    if err := binary.Write(w, binary.LittleEndian, &unix)  ; err != nil { return err }

    /* write the code */
    if n, err := w.Write(code); err != nil {
        return err
    } else if n != len(code) {
        return io.ErrShortWrite
    }

    /* add paddings to ensure file is larger than 4k */
    if d := 4096 - int(hlen + clen); d <= 0 {
        return nil
    } else if n, err := w.Write(zeroBytes[:d]); err != nil {
        return err
    } else if n != d {
        return io.ErrShortWrite
    } else {
        return nil
    }
}
