package obj

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

type CommandHeader struct {
    Cmd  uint32
    Size uint32
}

type SegmentCommand struct {
    CommandHeader
    SegmentName  [16]byte
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

type UnixThreadHeader struct {
    CommandHeader
    Flavor uint32
    Count  uint32
    Reg    Registers
}
