package obj

import (
    `io`
)

// Format represents the saved binary file format.
type Format int

const (
    // FmtELF indicates it should be saved as an ELF binary.
    FmtELF Format = iota + 1

    // FmtMachO indicates it should be saved as a Mach-O binary.
    FmtMachO
)

// BinaryWriter represents a writer that saves assembled
// machine code into binary executable file.
type BinaryWriter interface {
    Save(buf *io.Writer, code []byte) error
}
