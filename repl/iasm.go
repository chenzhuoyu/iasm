package repl

import (
    `fmt`
    `io`
    `math`
    `os`
    `path`
    `runtime`
    `strings`
    `unicode`
    `unsafe`

    `github.com/knz/go-libedit`
)

type IASM struct {
    run bool
    mem map[uint64]_Memory
    fns map[string]unsafe.Pointer
}

func CreateIASM() *IASM {
    return &IASM {
        run: true,
        mem: map[uint64]_Memory{},
        fns: map[string]unsafe.Pointer{},
    }
}

var (
    _HistoryFile = path.Clean(os.ExpandEnv("$HOME/.iasmhistory"))
)

// Start starts a new REPL session.
func (self *IASM) Start() {
    var err error
    var cmd string
    var ret libedit.EditLine

    /* initialize libedit */
    if ret, err = libedit.Init("iasm", false); err != nil {
        panic(err)
    }

    /* initialze the editor instance */
    defer ret.Close()
    ret.RebindControlKeys()

    /* enable history */
    _ = ret.UseHistory(-1, true)
    _ = ret.LoadHistory(_HistoryFile)

    /* set prompt and enable auto-save */
    ret.SetLeftPrompt(">>> ")
    ret.SetAutoSaveHistory(_HistoryFile, true)

    /* greeding prompts */
    println("Interactive Assembler v1.0")
    println("Compiled with [" + runtime.Version() + "].")
    println("Loading history from '" + _HistoryFile + "'.")
    println(`Type ".help" for more information.`)

    /* main REPL loop */
    for self.run {
        if cmd, err = ret.GetLine(); err == nil {
            self.handleCommand(&ret, strings.TrimSpace(cmd))
        } else if err == io.EOF {
            self.handleEOF()
        } else if err == libedit.ErrInterrupted {
            println("^C")
        } else {
            panic(err)
        }
    }
}

func (self *IASM) handleEOF() {
    self.run = false
    println()
}

func (self *IASM) handleCommand(el *libedit.EditLine, cmd string) {
    var pos int
    var fnv func(*IASM, string)

    /* check for empty command */
    if cmd == "" {
        return
    }

    /* add the command to history */
    if err := el.AddHistory(cmd); err != nil {
        panic(err)
    }

    /* find the command delimiter */
    if pos = strings.IndexFunc(cmd, unicode.IsSpace); pos == -1 {
        pos = len(cmd)
    }

    /* handle syntax errors via panic/recover */
    defer func() {
        if v := recover(); v != nil {
            if e, ok := v.(_SyntaxError); !ok {
                panic(v)
            } else {
                println("iasm: " + e.Error())
            }
        }
    }()

    /* call the command, if any */
    if fnv = _CMDS[cmd[:pos]]; fnv != nil {
        fnv(self, strings.TrimSpace(cmd[pos:]))
    } else {
        println("iasm: unknown command: " + cmd)
    }
}

var _CMDS = map[string]func(*IASM, string) {
    ".free"   : (*IASM)._cmd_free,
    ".malloc" : (*IASM)._cmd_malloc,
    ".info"   : (*IASM)._cmd_info,
    ".read"   : (*IASM)._cmd_read,
    ".write"  : (*IASM)._cmd_write,
    ".fill"   : (*IASM)._cmd_fill,
    ".regs"   : (*IASM)._cmd_regs,
    ".exit"   : (*IASM)._cmd_exit,
    ".help"   : (*IASM)._cmd_help,
}

func (self *IASM) _cmd_free(v string) {
    var ok  bool
    var err error
    var mid uint64
    var mem _Memory

    /* parse the memory ID */
    scan  (v).
    uint  (&mid).
    close ()

    /* find the memory block */
    if mem, ok = self.mem[mid]; !ok {
        println(fmt.Sprintf("iasm: no such memory block with ID %d", mid))
        return
    }

    /* unmap the memory block */
    if err = munmap(mem); err == nil {
        delete(self.mem, mid)
    } else {
        println("iasm: munmap(): " + err.Error())
    }
}

func (self *IASM) _cmd_malloc(v string) {
    var err error
    var mid uint64
    var nbs uint64
    var mem _Memory

    /* parse the memory ID and size */
    scan  (v).
    uint  (&mid).
    uint  (&nbs).
    close ()

    /* check for duplicaded IDs */
    if _, ok := self.mem[mid]; ok {
        println(fmt.Sprintf("iasm: duplicated memory ID: %d", mid))
        return
    }

    /* map the memory */
    if mem, err = mmap(nbs); err != nil {
        println("iasm: mmap(): " + err.Error())
        return
    }

    /* save the memory block */
    self.mem[mid] = mem
    println(fmt.Sprintf("Memory mapped at address %#x with size %d", mem.addr, mem.size))
}

func (self *IASM) _cmd_info(v string) {
    var ok  bool
    var mid uint64
    var mem _Memory

    /* parse the memory ID */
    scan  (v).
    uint  (&mid).
    close ()

    /* find the memory block */
    if mem, ok = self.mem[mid]; !ok {
        println(fmt.Sprintf("iasm: no such memory block with ID %d", mid))
        return
    }

    /* print the memory info */
    println(fmt.Sprintf("Address : %#x", mem.addr))
    println(fmt.Sprintf("Size    : %d bytes", mem.size))
}

func (self *IASM) _cmd_read(v string) {
    var ok  bool
    var off uint64
    var mid uint64
    var mem _Memory

    /* parse the memory ID, offset and the optional size */
    nbs := uint64(math.MaxUint64)
    scan(v).idoff(&mid, &off).uintopt(&nbs).close()

    /* find the memory block */
    if mem, ok = self.mem[mid]; !ok {
        println(fmt.Sprintf("iasm: no such memory block with ID %d", mid))
        return
    }

    /* read the whole block if not specified */
    if nbs == math.MaxUint64 {
        nbs = mem.size - off
    }

    /* check the offset and size */
    if off + nbs > mem.size {
        if diff := off + nbs - mem.size; diff == 1 {
            println("iasm: memory read 1 byte past the boundary")
            return
        } else {
            println(fmt.Sprintf("iasm: memory read %d bytes past the boundary", diff))
            return
        }
    }

    /* dump the content */
    print(hexdump(
        mem.buf()[off:off + nbs],
        mem.addr,
    ))
}

func (self *IASM) _cmd_write(v string) {
    var ok  bool
    var val string
    var off uint64
    var mid uint64
    var mem _Memory

    /* parse the memory ID, offset, filling byte and the optional size */
    scan(v).idoff(&mid, &off).str(&val).close()
    nbs := uint64(len(val))

    /* find the memory block */
    if mem, ok = self.mem[mid]; !ok {
        println(fmt.Sprintf("iasm: no such memory block with ID %d", mid))
        return
    }

    /* check the offset and size */
    if off + nbs > mem.size {
        if diff := off + nbs - mem.size; diff == 1 {
            println("iasm: memory fill 1 byte past the boundary")
        } else {
            println(fmt.Sprintf("iasm: memory fill %d bytes past the boundary", diff))
        }
    }

    /* copy the data into memory */
    copy(mem.buf()[off:], val)
    println(fmt.Sprintf("%d bytes written into %#x+%#x", nbs, mem.addr, off))
}

func (self *IASM) _cmd_fill(v string) {
    var ok  bool
    var val uint64
    var off uint64
    var mid uint64
    var mem _Memory

    /* parse the memory ID, offset, filling byte and the optional size */
    nbs := uint64(math.MaxUint64)
    scan(v).idoff(&mid, &off).uint(&val).uintopt(&nbs).close()

    /* check the filling value */
    if val > 255 {
        println(fmt.Sprintf("iasm: invalid filling value: %d is not a byte", val))
        return
    }

    /* find the memory block */
    if mem, ok = self.mem[mid]; !ok {
        println(fmt.Sprintf("iasm: no such memory block with ID %d", mid))
        return
    }

    /* read the whole block if not specified */
    if nbs == math.MaxUint64 {
        nbs = mem.size - off
    }

    /* check the offset and size */
    if off + nbs > mem.size {
        if diff := off + nbs - mem.size; diff == 1 {
            println("iasm: memory fill 1 byte past the boundary")
        } else {
            println(fmt.Sprintf("iasm: memory fill %d bytes past the boundary", diff))
        }
    }

    /* fill the memory with byte */
    for i := off; i < off + nbs; i++ {
        *(*byte)(unsafe.Pointer(uintptr(mem.p()) + uintptr(i))) = byte(val)
    }
}

func (self *IASM) _cmd_regs(v string) {
    regs := _regs.Dump(13)
    sels := map[string]bool{}

    /* check for register selectors */
    if fv := strings.Fields(v); len(fv) != 0 {
        for _, x := range fv {
            sels[strings.ToLower(x)] = true
        }
    }

    /* dump the registers */
    for _, reg := range regs {
        if v == "*" || sels[reg.reg] || (!reg.vec && len(sels) == 0) {
            println(fmt.Sprintf("%10s = %s", reg.reg, reg.val))
        }
    }
}

func (self *IASM) _cmd_exit(_ string) {
    self.run = false
}

func (self *IASM) _cmd_help(_ string) {
    println("Supported commands:")
    println("    .free   ID ........................ Free a block of memory with ID.")
    println("    .malloc ID SIZE ................... Allocate a block of memory with ID of")
    println("                                        SIZE bytes.")
    println()
    println("    .info   ID ........................ Print basic informations of a memory")
    println("                                        block identified by ID.")
    println()
    println("    .read   ID[+OFF] [SIZE] ........... Read a block of memory identified by")
    println("                                        ID[+OFF] of SIZE bytes, default to the")
    println("                                        whole block if SIZE is not set.")
    println()
    println("    .write  ID[+OFF] DATA ............. Write DATA into a block of memory")
    println("                                        identified by ID[+OFF].")
    println()
    println("    .fill   ID[+OFF] BYTE [SIZE] ...... Fill a block of memory identified by")
    println("                                        ID[+OFF] with BYTE of SIZE bytes,")
    println("                                        default to the whole block if SIZE is")
    println("                                        not set.")
    println()
    println("    .regs   [REG*] .................... Print the content of the specified")
    println("                                        registers, default to general purpose")
    println("                                        registers if not specified. To also")
    println(`                                        include SIMD registers, type ".regs *".`)
    println()
    println("    .new    NAME ...................... Create a new function with NAME.")
    println("    .run    NAME ...................... Execute the function NAME.")
    println("    .unload NAME ...................... Unload and delete the function NAME.")
    println("    .exit ............................. Exit Interactive Assembler.")
    println("    .help ............................. This help message.")
    println()
    println("You can also execute assembly instructions of the current CPU architecture")
    println("directly in this prompt without creating a function every time, by typing the")
    println("instruction and then press ENTER.")
    println()
}
