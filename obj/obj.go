package obj

import (
    `fmt`
    `io`
    `os`
)

// Format represents the saved binary file format.
type Format int

const (
    // ELF indicates it should be saved as an ELF binary.
    ELF Format = iota

    // MachO indicates it should be saved as a Mach-O binary.
    MachO
)

var formatTab = [...]func(w io.Writer, code []byte) error {
    ELF   : nil,
    MachO : assembleMachO,
}

var formatNames = map[Format]string {
    ELF   : "ELF",
    MachO : "Mach-O",
}

// String returns the name of a specified format.
func (self Format) String() string {
    if v, ok := formatNames[self]; ok {
        return v
    } else {
        return fmt.Sprintf("Format(%d)", self)
    }
}

// Write assembles a binary executable.
func (self Format) Write(w io.Writer, code []byte) error {
    if self >= 0 && int(self) < len(formatTab) && formatTab[self] != nil {
        return formatTab[self](w, code)
    } else {
        return fmt.Errorf("unsupported format: %s", self)
    }
}

// Generate generates a binary executable file from the specified code.
func (self Format) Generate(name string, code []byte) error {
    var fp *os.File
    var err error

    /* create the output file */
    if fp, err = os.Create(name); err != nil {
        return err
    }

    /* generate the code */
    if err = self.Write(fp, code); err != nil {
        _ = fp.Close()
        _ = os.Remove(name)
        return err
    }

    /* close the file and make it executable */
    _ = fp.Close()
    _ = os.Chmod(name, 0755)
    return nil
}
