package main

import (
    `os`

    `github.com/chenzhuoyu/iasm/obj`
    `github.com/chenzhuoyu/iasm/repl`
    `github.com/chenzhuoyu/iasm/x86_64`
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
    println("       iasm -h | --help")
    println()
    println("General Options:")
    println(`    -D DEF, --define=DEF        Passing the defination to preprocessor`)
    println("    -f FMT, --format=FMT        Select output format")
    println("       bin                          Flat raw binary (default)")
    println("       macho                        Mach-O executable")
    println("       elf                          ELF executable")
    println()
    println("    -h, --help                  This help message")
    println("    -o FILE, --output=FILE      Output file name")
    println("    -s, --gas-compat            GAS compatible mode")
    println()
    println("Environment Variables:")
    println("    CPP                         The C Preprocessor")
    println()
}

func compile() {
    var err error
    var src string
    var rem []string
    var asm x86_64.Assembler
    var ret []optparse.Result

    /* options list */
    opts := []optparse.Option {
        { "help"       , 'h', optparse.KindNone     },
        { "define"     , 'D', optparse.KindRequired },
        { "format"     , 'f', optparse.KindRequired },
        { "output"     , 'o', optparse.KindRequired },
        { "gas-compat" , 's', optparse.KindNone     },
    }

    /* parse the options */
    if ret, rem, err = optparse.Parse(opts, os.Args); err != nil {
        println("iasm: error: " + err.Error())
        usage()
    }

    /* default values */
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
    }

    /* must have source files */
    if len(rem) == 0 {
        println("iasm: error: missing input file.")
        os.Exit(1)
    }

    /* must have exactly 1 source file */
    if len(rem) != 1 {
        println("iasm: error: too many input files.")
        os.Exit(1)
    }

    /* preprocess the source file */
    if src, err = preprocess(rem[0], defs); err != nil {
        println("iasm: error: failed to run preprocessor: " + err.Error())
        os.Exit(1)
    }

    /* check for GAS compatible mode */
    if mgas {
        asm.Options().InstructionAliasing = true
        asm.Options().IgnoreUnknownDirectives = true
    }

    /* assemble the source */
    if err = asm.Assemble(src); err != nil {
        println("iasm: error: " + err.Error())
        os.Exit(1)
    }

    /* check for format */
    switch formatTab[ffmt] {
        case _F_bin   : err = os.WriteFile(fout, asm.Code(), 0755)
        case _F_elf   : err = obj.ELF.Generate(fout, asm.Code(), uint64(asm.Base()), uint64(asm.Entry()))
        case _F_macho : err = obj.MachO.Generate(fout, asm.Code(), uint64(asm.Base()), uint64(asm.Entry()))
        default       : panic("invalid format: " + ffmt)
    }

    /* check for errors */
    if err != nil {
        println("iasm: error: " + err.Error())
        os.Exit(1)
    }
}

func main() {
    if len(os.Args) != 1 {
        compile()
    } else {
        new(repl.IASM).Start()
    }
}
