package repl

import (
    `fmt`
    `io`
    `math`
    `os`
    `path`
    `runtime`
    `strconv`
    `strings`
    `unicode`
    `unsafe`

    `github.com/knz/go-libedit`
)

// IASM is the interactive REPL.
type IASM struct {
    run bool
    off uintptr
    efd libedit.EditLine
    ias _IASMArchSpecific
    mem map[uint64]_Memory
    fns map[string]unsafe.Pointer
}

// HistoryFile is the file that saves the REPL history, defaults to "~/.iasmhistory".
var HistoryFile = path.Clean(os.ExpandEnv("$HOME/.iasmhistory"))

// Start starts a new REPL session.
func (self *IASM) Start() {
    var err error
    var efd libedit.EditLine

    /* greeting messages */
    println("Interactive Assembler v1.0")
    println("Compiled with " + strconv.Quote(runtime.Version()) + ".")
    println("History will be loaded from " + strconv.Quote(HistoryFile) + ".")
    println(`Type ".help" for more information.`)

    /* initialize libedit */
    if efd, err = libedit.Init("iasm", false); err != nil {
        panic(err)
    }

    /* initialize IASM */
    self.off = 0
    self.efd = efd
    self.run = true
    self.mem = map[uint64]_Memory{}
    self.fns = map[string]unsafe.Pointer{}

    /* initialze the editor instance */
    defer self.efd.Close()
    self.efd.RebindControlKeys()

    /* enable history */
    _ = self.efd.UseHistory(-1, true)
    _ = self.efd.LoadHistory(HistoryFile)

    /* set prompt and enable auto-save */
    self.efd.SetLeftPrompt(">>> ")
    self.efd.SetAutoSaveHistory(HistoryFile, true)

    /* main REPL loop */
    for self.run {
        self.handleOnce()
    }
}

func (self *IASM) readLine() string {
    var err error
    var ret string

    /* read the line */
    for {
        if ret, err = self.efd.GetLine(); err == nil {
            break
        } else if err != libedit.ErrInterrupted {
            panic(err)
        } else {
            println("^C")
        }
    }

    /* check for empty line */
    if ret == "" {
        return ""
    }

    /* add to history */
    _ = self.efd.AddHistory(ret)
    return ret
}

func (self *IASM) handleEOF() {
    self.run = false
    println()
}

func (self *IASM) handleOnce() {
    defer self.handleError()
    self.handleCommand(strings.TrimSpace(self.readLine()))
}

func (self *IASM) handleError() {
    switch v := recover(); v {
        case nil    : break
        case io.EOF : self.handleEOF()
        default     : println(fmt.Sprintf("iasm: %v", v))
    }
}

func (self *IASM) handleCommand(cmd string) {
    var pos int
    var fnv func(*IASM, string)

    /* check for empty command */
    if cmd == "" {
        return
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

    /* assembly immediately or call the command, if any */
    if cmd[0] != '.' {
        self.handleAsmImmediate(cmd)
    } else if fnv = _CMDS[cmd[1:pos]]; fnv != nil {
        fnv(self, strings.TrimSpace(cmd[pos:]))
    } else {
        println("iasm: unknown command: " + cmd)
    }
}

func (self *IASM) handleAsmImmediate(asm string) {
    var err error
    var buf []byte

    /* assemble the instruction */
    if buf, err = self.ias.doasm(self.off, asm); err != nil {
        println("iasm: " + err.Error())
        return
    }

    /* dump the instruction, and adjust the display offset */
    println(asmdump(buf, self.off, asm))
    self.off += uintptr(len(buf))
}

var _CMDS = map[string]func(*IASM, string) {
    "free"   : (*IASM)._cmd_free,
    "malloc" : (*IASM)._cmd_malloc,
    "info"   : (*IASM)._cmd_info,
    "read"   : (*IASM)._cmd_read,
    "write"  : (*IASM)._cmd_write,
    "fill"   : (*IASM)._cmd_fill,
    "regs"   : (*IASM)._cmd_regs,
    "asm"    : (*IASM)._cmd_asm,
    "sys"    : (*IASM)._cmd_sys,
    "base"   : (*IASM)._cmd_base,
    "exit"   : (*IASM)._cmd_exit,
    "help"   : (*IASM)._cmd_help,
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
    scan    (v).
    uint    (&mid).
    uintopt (&nbs).
    close   ()

    /* default to one page of memory */
    if nbs == 0 {
        nbs = _PageSize
    }

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

func (self *IASM) _cmd_asm(v string) {
    var ok  bool
    var err error
    var off uint64
    var mid uint64
    var fnv uintptr
    var mem _Memory

    /* parse the memory ID, offset, filling byte and the optional size */
    scan  (v).
    idoff (&mid, &off).
    close ()

    /* find the memory block */
    if mem, ok = self.mem[mid]; !ok {
        println(fmt.Sprintf("iasm: no such memory block with ID %d", mid))
        return
    }

    /* check for memory boundary */
    if fnv = mem.addr + uintptr(off); off >= mem.size {
        println("iasm: indexing past the end of memory")
        return
    }

    /* prompt messages */
    println(fmt.Sprintf("Assemble in memory block #(%d)+%#x (%#x).", mid, off, fnv))
    println(`Type ".end" and ENTER to end the assembly session.`)

    /* restore prompt later */
    buf := []byte(nil)
    rem := mem.size - off
    defer self.efd.SetLeftPrompt(">>> ")
    self.efd.SetLeftPrompt(fmt.Sprintf("(%#x) ", fnv))

    /* read and assemble the assembly source */
    for {
        src := strings.TrimSuffix(self.readLine(), "\n")
        val := strings.TrimSpace(src)

        /* check for end of assembly */
        if val == ".end" {
            break
        }

        /* feed into the assembler */
        if buf, err = self.ias.doasm(fnv, src); err != nil {
            println("iasm: assembly failed: " + err.Error())
            continue
        }

        /* check for memory space */
        if rem < uint64(len(buf)) {
            println(fmt.Sprintf("iasm: no space left in memory block: %d < %d", rem, len(buf)))
            continue
        }

        /* update the input line if stdout is a terminal */
        if isatty(os.Stdout.Fd()) {
            println("\x1b[F\x1b[K" + asmdump(buf, fnv, src))
        }

        /* save the machine code */
        ptr := mem.buf()
        copy(ptr[off:], buf)

        /* update the prompt and offsets */
        rem -= uint64(len(buf))
        off += uint64(len(buf))
        fnv += uintptr(len(buf))
        self.efd.SetLeftPrompt(fmt.Sprintf("(%#x) ", fnv))
    }
}

func (self *IASM) _cmd_sys(v string) {
    var ok  bool
    var err error
    var off uint64
    var mid uint64
    var fnv uintptr
    var mem _Memory
    var rs0 _RegFile
    var rs1 _RegFile

    /* parse the memory ID, offset, filling byte and the optional size */
    scan  (v).
    idoff (&mid, &off).
    close ()

    /* find the memory block */
    if mem, ok = self.mem[mid]; !ok {
        println(fmt.Sprintf("iasm: no such memory block with ID %d", mid))
        return
    }

    /* check for memory boundary */
    if fnv = mem.addr + uintptr(off); off >= mem.size {
        println("iasm: indexing past the end of memory")
        return
    }

    /* execute the code */
    if rs0, rs1, err = _exec.Execute(fnv); err != nil {
        println(fmt.Sprintf("iasm: cannot execute at memory address %#x: %s", fnv, err))
        return
    }

    /* print the differences */
    for _, diff := range rs0.Compare(rs1, 13) {
        println(fmt.Sprintf("%10s = %s", diff.reg, diff.val))
    }
}

func (self *IASM) _cmd_base(v string) {
    scan(v).uint((*uint64)(unsafe.Pointer(&self.off))).close()
}

func (self *IASM) _cmd_exit(_ string) {
    self.run = false
}

func (self *IASM) _cmd_help(_ string) {
    println("Supported commands:")
    println("    .free   ID ........................ Free a block of memory with ID.")
    println()
    println("    .malloc ID [SIZE] ................. Allocate a block of memory with ID of")
    println("                                        SIZE bytes if specified, or one page of")
    println("                                        memory if SIZE is not present.")
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
    println("    .asm    ID[+OFF] .................. Assemble into memory block identified by")
    println("                                        ID[+OFF].")
    println()
    println("    .sys    ID[+OFF] .................. Execute code in memory block identified")
    println("                                        by ID[+OFF] with the CALL instruction.")
    println()
    println("    .base   [BASE] .................... Set the base address for immediate")
    println("                                        assembling mode, just for display.")
    println()
    println("    .exit ............................. Exit Interactive Assembler.")
    println("    .help ............................. This help message.")
    println()
}
