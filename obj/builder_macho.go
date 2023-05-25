package obj

import (
    `debug/macho`
    `encoding/binary`
    `fmt`
    `io`
    `math/bits`
    `syscall`
    `unsafe`
)

const (
    _N_EXT  = 0x01
    _N_SECT = 0x0e
)

const (
    _S_ATTR_SOME_INSTRUCTIONS = 0x00000400
    _S_ATTR_PURE_INSTRUCTIONS = 0x80000000
)

const (
    _ENTRY_SYM = "_main"
)

type _MachHeader64 struct {
    Magic  uint32
    Cpu    macho.Cpu
    SubCpu uint32
    Type   macho.Type
    Ncmd   uint32
    Cmdsz  uint32
    Flags  uint32
    _      uint32
}

type _ObjectFileHeader struct {
    mach     _MachHeader64
    textseg  macho.Segment64
    textsect macho.Section64
    symtab   macho.SymtabCmd
    dsymtab  macho.DysymtabCmd
}

func buildMachO(w io.Writer, arch *Arch, code []byte, base uintptr, entry uintptr) error {
    var n int
    var err error

    /* check for base address alignment */
    if base & (arch.Align - 1) != 0 {
        return fmt.Errorf("base address must be aligned to %d bytes", arch.Align)
    }

    /* check for entry point alignment */
    if entry & (arch.Align - 1) != 0 {
        return fmt.Errorf("entry point must be aligned to %d bytes", arch.Align)
    }

    /* calculate symbol table & strings offset */
    symoff  := unsafe.Sizeof(_ObjectFileHeader{}) + uintptr(len(code))
    symbase := (((symoff - 1) >> 4) + 1) << 4

    /* machO object header */
    hdr := _ObjectFileHeader {
        mach: _MachHeader64 {
            Magic  : macho.Magic64,
            Cpu    : arch.MachOCpu,
            SubCpu : arch.MachOSubCpu,
            Type   : macho.TypeObj,
            Ncmd   : 3,
            Cmdsz  : uint32(unsafe.Sizeof(_ObjectFileHeader{}) - unsafe.Sizeof(_MachHeader64{})),
            Flags  : macho.FlagNoUndefs,
        },
        textseg: macho.Segment64 {
            Cmd     : macho.LoadCmdSegment64,
            Len     : uint32(unsafe.Sizeof(macho.Segment64{}) + unsafe.Sizeof(macho.Section64{})),
            Name    : [16]byte{},
            Addr    : uint64(base),
            Memsz   : uint64(len(code)),
            Offset  : uint64(unsafe.Sizeof(_ObjectFileHeader{})),
            Filesz  : uint64(len(code)),
            Maxprot : syscall.PROT_READ | syscall.PROT_EXEC,
            Prot    : syscall.PROT_READ | syscall.PROT_EXEC,
            Nsect   : 1,
            Flag    : 0,
        },
        textsect: macho.Section64 {
            Name   : str16b("__text"),
            Seg    : str16b("__TEXT"),
            Addr   : uint64(base),
            Size   : uint64(len(code)),
            Offset : uint32(unsafe.Sizeof(_ObjectFileHeader{})),
            Align  : uint32(bits.TrailingZeros64(uint64(arch.Align))),
            Reloff : 0,
            Nreloc : 0,
            Flags  : _S_ATTR_SOME_INSTRUCTIONS | _S_ATTR_PURE_INSTRUCTIONS,
        },
        symtab: macho.SymtabCmd {
            Cmd     : macho.LoadCmdSymtab,
            Len     : uint32(unsafe.Sizeof(macho.SymtabCmd{})),
            Symoff  : uint32(symbase),
            Nsyms   : 1,
            Stroff  : uint32(symbase + unsafe.Sizeof(macho.Nlist64{})),
            Strsize : uint32(len(_ENTRY_SYM) + 1),
        },
        dsymtab: macho.DysymtabCmd {
            Cmd            : macho.LoadCmdDysymtab,
            Len            : uint32(unsafe.Sizeof(macho.DysymtabCmd{})),
            Ilocalsym      : 0,
            Nlocalsym      : 0,
            Iextdefsym     : 0,
            Nextdefsym     : 1,
            Iundefsym      : 0,
            Nundefsym      : 0,
            Tocoffset      : 0,
            Ntoc           : 0,
            Modtaboff      : 0,
            Nmodtab        : 0,
            Extrefsymoff   : 0,
            Nextrefsyms    : 0,
            Indirectsymoff : 0,
            Nindirectsyms  : 0,
            Extreloff      : 0,
            Nextrel        : 0,
            Locreloff      : 0,
            Nlocrel        : 0,
        },
    }

    /* symbol table */
    tab := macho.Nlist64 {
        Name  : 1,
        Type  : _N_SECT | _N_EXT,
        Sect  : 1,
        Desc  : 0,
        Value : uint64(entry),
    }

    /* write the object file */
    if err = binary.Write(w, binary.LittleEndian, &hdr); err != nil {
        return err
    } else if n, err = w.Write(code); err != nil {
        return err
    } else if n != len(code) {
        return io.ErrShortWrite
    } else if err = binary.Write(w, binary.LittleEndian, make([]byte, symbase - symoff)); err != nil {
        return err
    } else if err = binary.Write(w, binary.LittleEndian, &tab); err != nil {
        return err
    } else if err = binary.Write(w, binary.LittleEndian, byte(0)); err != nil {
        return err
    } else if err = binary.Write(w, binary.LittleEndian, []byte(_ENTRY_SYM)); err != nil {
        return err
    } else if err = binary.Write(w, binary.LittleEndian, byte(0)); err != nil {
        return err
    } else {
        return nil
    }
}
