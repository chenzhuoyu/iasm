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

var _ArchTab = map[string]string {
    "arm64": "aarch64",
    "amd64": "x86_64",
}

func usage() {
    println("usage: iasm [OPTIONS] <source>")
    println("       iasm [OPTIONS] -i")
    println("       iasm -h | --help")
    println()
    println("General Options:")
    println(`    -a ARCH, --arch=ARCH        Target architecture`)
    println(`    -c, --compile               Compile only`)
    println(`    -D DEF, --define=DEF        Passing the defination to preprocessor`)
    println("    -h, --help                  This help message")
    println("    -i, --interactive           Start as an interactive REPL shell")
    println("    -o FILE, --output=FILE      Output file name")
    println("    -s, --gas-compat            GAS compatible mode")
    println()
    println("Environment Variables:")
    println("    CPP                         The C Preprocessor")
    println("    LD                          Path to the linker program")
    println()
    println("Supported Architectures:")
    println("    " + strings.Join(asm.SupportedArch(), "\n    "))
    println()
}

func main() {
    var err error
    var src string
    var rem []string
    var cpu *asm.Arch
    var dst *obj.Arch
    var out *asm.Assembler
    var ret []optparse.Result

    /* options list */
    opts := []optparse.Option {
        { "arch"        , 'a', optparse.KindRequired },
        { "compile"     , 'c', optparse.KindNone     },
        { "define"      , 'D', optparse.KindRequired },
        { "format"      , 'f', optparse.KindRequired },
        { "help"        , 'h', optparse.KindNone     },
        { "interactive" , 'i', optparse.KindNone     },
        { "output"      , 'o', optparse.KindRequired },
        { "gas-compat"  , 's', optparse.KindNone     },
    }

    /* parse the options */
    if ret, rem, err = optparse.Parse(opts, os.Args); err != nil {
        println("iasm: error: " + err.Error())
        usage()
        return
    }

    /* default values */
    arch := ""
    comp := false
    intr := false
    help := false
    mgas := false
    fout := "a.out"
    defs := []string(nil)

    /* check the result */
    for _, vv := range ret {
        switch vv.Short {
            case 'c': comp = true
            case 'h': help = true
            case 'i': intr = true
            case 's': mgas = true
            case 'a': arch = vv.Optarg
            case 'o': fout = vv.Optarg
            case 'D': defs = append(defs, vv.Optarg)
        }
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
    if cpu = asm.GetArch(arch); cpu == nil {
        println("iasm: unsupported architecture: " + arch)
        println("iasm: available architectures are: " + strings.Join(asm.SupportedArch(), ", "))
        os.Exit(1)
    }

    /* enter interactive mode if command line arguments or explicitly specified */
    if intr || len(os.Args) == 1 {
        new(repl.IASM).Start(cpu)
        return
    }

    /* find the object architecture */
    if dst = obj.GetArch(arch); dst == nil {
        panic("iasm: unimplemented object architecture: " + arch)
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
    if out = cpu.CreateAssembler(); mgas {
        out.Options().PermissiveMnemonic = true
        out.Options().IgnoreUnknownDirectives = true
    }

    /* assemble the source */
    if err = out.Assemble(src); err != nil {
        println("iasm: error: " + err.Error())
        os.Exit(1)
    }

    /* compile the machine code, link if required */
    if comp {
        err = obj.CurrentOS.Compile(fout, dst, out.Code(), out.Base(), out.Entry())
    } else {
        err = obj.CurrentOS.CompileAndLink(fout, dst, out.Code(), out.Base(), out.Entry())
    }

    /* check for errors */
    if err != nil {
        println("iasm: error: " + err.Error())
        os.Exit(1)
    }
}
