package main

import (
    `os`
    `runtime`
    `strings`

    _ `github.com/chenzhuoyu/iasm/arch/aarch64`
    _ `github.com/chenzhuoyu/iasm/arch/x86_64`
    `github.com/chenzhuoyu/iasm/asm`
    `github.com/chenzhuoyu/iasm/obj`
    `github.com/chenzhuoyu/iasm/repl`
    `nullprogram.com/x/optparse`
)

type _FileFormat int

const (
    _F_bin _FileFormat = iota + 1
    _F_macho
    _F_elf
)

var formatTab = map[string]_FileFormat {
    "bin"   : _F_bin,
    "macho" : _F_macho,
    "elf"   : _F_elf,
}

func usage() {
    println("usage: iasm [OPTIONS] <source>")
    println("       iasm [OPTIONS] -i")
    println("       iasm -h | --help")
    println()
    println("General Options:")
    println(`    -a ARCH, --arch=ARCH        Target architecture`)
    println(`    -D DEF, --define=DEF        Passing the defination to preprocessor`)
    println("    -f FMT, --format=FMT        Select output format")
    println("       bin                          Flat raw binary (default)")
    println("       macho                        Mach-O executable")
    println("       elf                          ELF executable")
    println()
    println("    -h, --help                  This help message")
    println("    -o FILE, --output=FILE      Output file name")
    println("    -s, --gas-compat            GAS compatible mode")
    println("    -i, --interactive           Start as an interactive REPL shell")
    println()
    println("Environment Variables:")
    println("    CPP                         The C Preprocessor")
    println()
    println("Supported Architectures:")
    println("    " + strings.Join(asm.SupportedArch(), "\n    "))
    println()
}

var _ArchTab = map[string]string {
    "arm64": "aarch64",
    "amd64": "x86_64",
}

func main() {
    var err error
    var src string
    var rem []string
    var arc *asm.Arch
    var out *asm.Assembler
    var ret []optparse.Result

    /* options list */
    opts := []optparse.Option {
        { "help"        , 'h', optparse.KindNone     },
        { "arch"        , 'a', optparse.KindRequired },
        { "define"      , 'D', optparse.KindRequired },
        { "format"      , 'f', optparse.KindRequired },
        { "output"      , 'o', optparse.KindRequired },
        { "gas-compat"  , 's', optparse.KindNone     },
        { "interactive" , 'i', optparse.KindNone     },
    }

    /* parse the options */
    if ret, rem, err = optparse.Parse(opts, os.Args); err != nil {
        println("iasm: error: " + err.Error())
        usage()
        return
    }

    /* default values */
    arch := ""
    intr := false
    help := false
    mgas := false
    ffmt := "bin"
    fout := "a.out"
    defs := []string(nil)

    /* check the result */
    for _, vv := range ret {
        switch vv.Short {
            case 'h': help = true
            case 's': mgas = true
            case 'i': intr = true
            case 'a': arch = vv.Optarg
            case 'f': ffmt = vv.Optarg
            case 'o': fout = vv.Optarg
            case 'D': defs = append(defs, vv.Optarg)
        }
    }

    /* check file format */
    if _, ok := formatTab[ffmt]; !ok {
        println("iasm: error: unknown file format: " + ffmt)
        os.Exit(1)
    }

    /* check for help */
    if help {
        usage()
        return
    }

    /* default to current architecture */
    if arch == "" {
        if arch = _ArchTab[runtime.GOARCH]; arch == "" {
            arch = "x86_64"
            println(`iasm: unsupported architecture "` + runtime.GOARCH + `", use "x86_64" instead.`)
        }
    }

    /* create the architecture */
    if arc = asm.GetArch(arch); arc == nil {
        println("iasm: unsupported architecture: " + arch)
        println("iasm: available architectures are: " + strings.Join(asm.SupportedArch(), ", "))
        os.Exit(1)
    }

    /* enter interactive mode if command line arguments or explicitly specified */
    if intr || len(os.Args) == 1 {
        new(repl.IASM).Start(arc)
        return
    }

    /* check four source file */
    if len(rem) == 0 {
        println("iasm: error: missing input file.")
        os.Exit(1)
    }

    /* must have exactly 1 source file for compilation mode */
    if len(rem) != 1 {
        println("iasm: error: too many input files.")
        os.Exit(1)
    }

    /* preprocess the source file */
    if src, err = preprocess(rem[0], defs); err != nil {
        println("iasm: error: failed to run preprocessor: " + err.Error())
        os.Exit(1)
    }

    /* create the assembler, check for GAS compatible mode */
    if out = arc.CreateAssembler(); mgas {
        out.Options().PermissiveMnemonic = true
        out.Options().IgnoreUnknownDirectives = true
    }

    /* assemble the source */
    if err = out.Assemble(src); err != nil {
        println("iasm: error: " + err.Error())
        os.Exit(1)
    }

    /* check for format */
    switch formatTab[ffmt] {
        case _F_bin   : err = os.WriteFile(fout, out.Code(), 0755)
        case _F_elf   : err = obj.ELF.Generate(fout, out.Code(), uint64(out.Base()), uint64(out.Entry()))
        case _F_macho : err = obj.MachO.Generate(fout, out.Code(), uint64(out.Base()), uint64(out.Entry()))
        default       : panic("invalid format: " + ffmt)
    }

    /* check for errors */
    if err != nil {
        println("iasm: error: " + err.Error())
        os.Exit(1)
    }
}
