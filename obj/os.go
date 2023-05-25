package obj

import (
    `errors`
    `io`
    `os`
)

// OS reprensets the target operating system.
type OS uint8

const (
    Unsupported OS = iota
    MacOS
    Linux
)

var builderTab = [256]func(io.Writer, *Arch, []byte, uintptr, uintptr) error {
    MacOS: buildMachO,
    Linux: buildELF,
}

func (self OS) String() string {
    switch self {
        case MacOS : return "macOS"
        case Linux : return "Linux"
        default    : return "???"
    }
}

func (self OS) build(fp *os.File, arch *Arch, code []byte, base uintptr, entry uintptr) error {
    err := error(nil)
    file := fp.Name()

    /* compile the code */
    if builderTab[self] != nil {
        err = buildMachO(fp, arch, code, base, entry)
    } else {
        err = errors.New("unsupported operating system")
    }

    /* close the file, check for errors */
    if _ = fp.Close(); err == nil {
        return nil
    }

    /* remove the file when error occures */
    _ = os.Remove(file)
    return err
}

// Compile wraps the machine code into arch-specific object file.
func (self OS) Compile(name string, arch *Arch, code []byte, base uintptr, entry uintptr) error {
    if fp, err := os.Create(name); err != nil {
        return err
    } else if err = self.build(fp, arch, code, base, entry); err != nil {
        return err
    } else {
        return nil
    }
}

// CompileAndLink is like Compile, but also links the generated object file into an executable.
func (self OS) CompileAndLink(name string, arch *Arch, code []byte, base uintptr, entry uintptr) error {
    var fp *os.File
    var err error

    /* can only link for target OS */
    if self != CurrentOS {
        return errors.New("cannot link for other operating systems")
    }

    /* create a temporary file for object code */
    if fp, err = os.CreateTemp("", "iasm-object-*"); err != nil {
        return err
    }

    /* compile the code */
    if err = self.build(fp, arch, code, base, entry); err != nil {
        return err
    }

    /* link the target, and remove the intermediate file */
    err = link(name, fp.Name())
    _ = os.Remove(fp.Name())

    /* check for errors */
    if err == nil {
        return nil
    }

    /* remove the output file if error occures */
    _ = os.Remove(name)
    return err
}
