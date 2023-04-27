package asm

import (
    `bytes`
    `errors`
    `fmt`
    `math`
    `strconv`

    `github.com/chenzhuoyu/iasm/expr`
)

// Symbols is the symbol table used by Assembler.
type Symbols struct {
    pc    uintptr
    terms map[string]expr.Term
}

// Pos returns the current PC for the value of "."
func (self *Symbols) Pos() int64 {
    return int64(self.pc)
}

// Get resolves a symbol and returns it's value.
func (self *Symbols) Get(name string) (expr.Term, error) {
    if ret, ok := self.terms[name]; ok {
        return ret, nil
    } else {
        return nil, errors.New("undefined name: " + name)
    }
}

// Label resolves a label and returns its value, or creates a new one if not exists.
func (self *Symbols) Label(name string) (*Label, error) {
    var ok bool
    var lb *Label
    var tr expr.Term

    /* check for existing terms */
    if tr, ok = self.terms[name]; ok {
        if lb, ok = tr.(*Label); ok {
            return lb, nil
        } else {
            return nil, errors.New("name is not a label: " + name)
        }
    }

    /* create a new one as needed */
    lb = new(Label)
    lb.Name = name

    /* create the map if needed */
    if self.terms == nil {
        self.terms = make(map[string]expr.Term, 1)
    }

    /* register the label */
    self.terms[name] = lb
    return lb, nil
}

// Define defines a new symbol with name and term.
func (self *Symbols) Define(name string, term expr.Term) {
    var ok bool
    var tr expr.Term

    /* create the map if needed */
    if self.terms == nil {
        self.terms = make(map[string]expr.Term, 1)
    }

    /* check for existing terms */
    if tr, ok = self.terms[name]; !ok {
        self.terms[name] = term
    } else if _, ok = tr.(*Label); !ok {
        self.terms[name] = term
    } else {
        panic("conflicting term types: " + name)
    }
}

// Command describes an assembler command.
//
// The Command.Args describes both the arity and argument type with characters,
// the length is the number of arguments, the character itself represents the
// argument type.
//
// Possible values are:
//
//      s   This argument should be a string
//      e   This argument should be an expression
//      ?   The next argument is optional, and must be the last argument.
//
type Command struct {
    Args    string
    Handler func(*Assembler, *Program, []ParsedCommandArg) error
}

// Options controls the behavior of Assembler.
type Options struct {
    // PermissiveMnemonic instructs the Assembler to be permissive about mnemonic names.
    // Set to true and the Assembler will try harder to find instructions.
    PermissiveMnemonic bool

    // IgnoreUnknownDirectives specifies whether to report errors when encountered unknown directives.
    // Set to true ignores all unknwon directives silently, useful for parsing generated assembly.
    IgnoreUnknownDirectives bool
}

// Assembler assembles the entire assembly program and generates the corresponding
// machine code representations.
type Assembler struct {
    cc   int
    buf  []byte
    arch *Arch
    main string
    opts Options
    repo Symbols
    line *ParsedLine
}

var asmCommands = map[string]Command {
    "org"     : { "e"   , (*Assembler).assembleCommandOrg     },
    "set"     : { "ee"  , (*Assembler).assembleCommandSet     },
    "byte"    : { "e"   , (*Assembler).assembleCommandByte    },
    "word"    : { "e"   , (*Assembler).assembleCommandWord    },
    "long"    : { "e"   , (*Assembler).assembleCommandLong    },
    "quad"    : { "e"   , (*Assembler).assembleCommandQuad    },
    "fill"    : { "e?e" , (*Assembler).assembleCommandFill    },
    "space"   : { "e?e" , (*Assembler).assembleCommandFill    },
    "align"   : { "e?e" , (*Assembler).assembleCommandAlign   },
    "entry"   : { "e"   , (*Assembler).assembleCommandEntry   },
    "ascii"   : { "s"   , (*Assembler).assembleCommandAscii   },
    "asciz"   : { "s"   , (*Assembler).assembleCommandAsciz   },
    "p2align" : { "e?e" , (*Assembler).assembleCommandP2Align },
}

func (self *Assembler) init(arch *Arch) *Assembler {
    self.arch = arch
    return self
}

func (self *Assembler) checkArgs(i int, n int, v *ParsedCommand, isString bool) error {
    if i >= len(v.Args) {
        return self.Error(fmt.Sprintf("command %s takes exact %d arguments", strconv.Quote(v.Cmd), n))
    } else if isString && !v.Args[i].IsString {
        return self.Error(fmt.Sprintf("argument %d of command %s must be a string", i + 1, strconv.Quote(v.Cmd)))
    } else if !isString && v.Args[i].IsString {
        return self.Error(fmt.Sprintf("argument %d of command %s must be an expression", i + 1, strconv.Quote(v.Cmd)))
    } else {
        return nil
    }
}

func (self *Assembler) buildLabel(p *Program, lb *ParsedLabel) error {
    if v, err := self.repo.Label(lb.Name); err != nil {
        return err
    } else {
        p.Link(v)
        return nil
    }
}

func (self *Assembler) buildCommand(p *Program, line *ParsedCommand) error {
    var iv int
    var cc rune
    var ok bool
    var va bool
    var fn Command

    /* find the command */
    if fn, ok = asmCommands[line.Cmd]; !ok {
        if self.opts.IgnoreUnknownDirectives {
            return nil
        } else {
            return self.Error("no such command: " + strconv.Quote(line.Cmd))
        }
    }

    /* expected & real argument count */
    argx := len(fn.Args)
    argc := len(line.Args)

    /* check the arguments */
    loop: for iv, cc = range fn.Args {
        switch cc {
            case '?' : va = true; break loop
            case 's' : if err := self.checkArgs(iv, argx, line, true)  ; err != nil { return err }
            case 'e' : if err := self.checkArgs(iv, argx, line, false) ; err != nil { return err }
            default  : panic("invalid argument descriptor: " + strconv.Quote(fn.Args))
        }
    }

    /* simple case: non-variadic command */
    if !va {
        if argc == argx {
            return fn.Handler(self, p, line.Args)
        } else {
            return self.Error(fmt.Sprintf("command %s takes exact %d arguments", strconv.Quote(line.Cmd), argx))
        }
    }

    /* check for the descriptor */
    if iv != argx - 2 {
        panic("invalid argument descriptor: " + strconv.Quote(fn.Args))
    }

    /* variadic command and the final optional argument is set */
    if argc == argx - 1 {
        switch fn.Args[argx - 1] {
            case 's' : if err := self.checkArgs(iv, -1, line, true)  ; err != nil { return err }
            case 'e' : if err := self.checkArgs(iv, -1, line, false) ; err != nil { return err }
            default  : panic("invalid argument descriptor: " + strconv.Quote(fn.Args))
        }
    }

    /* check argument count */
    if argc == argx - 1 || argc == argx - 2 {
        return fn.Handler(self, p, line.Args)
    } else {
        return self.Error(fmt.Sprintf("command %s takes %d or %d arguments", strconv.Quote(line.Cmd), argx - 2, argx - 1))
    }
}

func (self *Assembler) assembleCommandOrg(_ *Program, argv []ParsedCommandArg) error {
    var err error
    var val int64

    /* evaluate the expression */
    if val, err = expr.Eval(argv[0].Value, nil); err != nil {
        return err
    }

    /* check for origin */
    if val < 0 {
        return self.Error(fmt.Sprintf("negative origin: %d", val))
    }

    /* ".org" must be the first command if any */
    if self.cc != 1 {
        return self.Error(".org must be the first command if present")
    }

    /* set the initial program counter */
    self.repo.pc = uintptr(val)
    return nil
}

func (self *Assembler) assembleCommandSet(_ *Program, argv []ParsedCommandArg) error {
    var err error
    var val *expr.Expr

    /* parse the expression */
    if val, err = expr.Parse(argv[1].Value, &self.repo); err != nil {
        return err
    }

    /* define the new identifier */
    self.repo.Define(argv[0].Value, val)
    return nil
}

func (self *Assembler) assembleCommandInt(p *Program, argv []ParsedCommandArg, fn func(*Program, *expr.Expr) *Instruction) error {
    var err error
    var val *expr.Expr

    /* parse the expression */
    if val, err = expr.Parse(argv[0].Value, &self.repo); err != nil {
        return err
    }

    /* add to the program */
    fn(p, val)
    return nil
}

func (self *Assembler) assembleCommandByte(p *Program, argv []ParsedCommandArg) error {
    return self.assembleCommandInt(p, argv, (*Program).Byte)
}

func (self *Assembler) assembleCommandWord(p *Program, argv []ParsedCommandArg) error {
    return self.assembleCommandInt(p, argv, (*Program).Word)
}

func (self *Assembler) assembleCommandLong(p *Program, argv []ParsedCommandArg) error {
    return self.assembleCommandInt(p, argv, (*Program).Long)
}

func (self *Assembler) assembleCommandQuad(p *Program, argv []ParsedCommandArg) error {
    return self.assembleCommandInt(p, argv, (*Program).Quad)
}

func (self *Assembler) assembleCommandFill(p *Program, argv []ParsedCommandArg) error {
    var fv byte
    var nb int64
    var ex error

    /* evaluate the size */
    if nb, ex = expr.Eval(argv[0].Value, nil); ex != nil {
        return ex
    }

    /* check for filling size */
    if nb < 0 {
        return self.Error(fmt.Sprintf("negative filling size: %d", nb))
    }

    /* check for optional filling value */
    if len(argv) == 2 {
        if val, err := expr.Eval(argv[1].Value, nil); err != nil {
            return err
        } else if val < math.MinInt8 || val > math.MaxUint8 {
            return self.Error(fmt.Sprintf("value %d cannot be represented with a byte", val))
        } else {
            fv = byte(val)
        }
    }

    /* fill with specified byte */
    p.Data(bytes.Repeat([]byte { fv }, int(nb)))
    return nil
}

func (self *Assembler) assembleCommandAlign(p *Program, argv []ParsedCommandArg) error {
    var nb int64
    var ex error
    var fv *expr.Expr

    /* evaluate the size */
    if nb, ex = expr.Eval(argv[0].Value, nil); ex != nil {
        return ex
    }

    /* check for alignment value */
    if nb <= 0 {
        return self.Error(fmt.Sprintf("zero or negative alignment: %d", nb))
    }

    /* alignment must be a power of 2 */
    if (nb & (nb - 1)) != 0 {
        return self.Error(fmt.Sprintf("alignment must be a power of 2: %d", nb))
    }

    /* check for optional filling value */
    if len(argv) == 2 {
        if v, err := expr.Parse(argv[1].Value, &self.repo); err == nil {
            fv = v
        } else {
            return err
        }
    }

    /* fill with specified byte, default to 0 if not specified */
    p.Align(uint64(nb), fv)
    return nil
}

func (self *Assembler) assembleCommandEntry(_ *Program, argv []ParsedCommandArg) error {
    name := argv[0].Value
    rbuf := []rune(name)

    /* check all the characters */
    for i, cc := range rbuf {
        if !isident0(cc) && (i == 0 || !isident(cc)) {
            return self.Error("entry point must be a label name")
        }
    }

    /* set the main entry point */
    self.main = name
    return nil
}

func (self *Assembler) assembleCommandAscii(p *Program, argv []ParsedCommandArg) error {
    p.Data([]byte(argv[0].Value))
    return nil
}

func (self *Assembler) assembleCommandAsciz(p *Program, argv []ParsedCommandArg) error {
    p.Data(append([]byte(argv[0].Value), 0))
    return nil
}

func (self *Assembler) assembleCommandP2Align(p *Program, argv []ParsedCommandArg) error {
    var nb int64
    var ex error
    var fv *expr.Expr

    /* evaluate the size */
    if nb, ex = expr.Eval(argv[0].Value, nil); ex != nil {
        return ex
    }

    /* check for alignment value */
    if nb <= 0 {
        return self.Error(fmt.Sprintf("zero or negative alignment: %d", nb))
    }

    /* check for optional filling value */
    if len(argv) == 2 {
        if v, err := expr.Parse(argv[1].Value, &self.repo); err == nil {
            fv = v
        } else {
            return err
        }
    }

    /* fill with specified byte, default to 0 if not specified */
    p.Align(1 << nb, fv)
    return nil
}

// Base returns the origin.
func (self *Assembler) Base() uintptr {
    return self.repo.pc
}

// Code returns the assembled machine code.
func (self *Assembler) Code() []byte {
    return self.buf
}

// Entry returns the address of the specified entry point, or the origin if not specified.
func (self *Assembler) Entry() uintptr {
    if self.main == "" {
        return self.repo.pc
    } else if tr, err := self.repo.Get(self.main); err != nil {
        panic(err)
    } else if val, err := tr.Evaluate(); err != nil {
        panic(err)
    } else {
        return uintptr(val)
    }
}

// Options returns the internal options reference, changing it WILL affect this Assembler instance.
func (self *Assembler) Options() *Options {
    return &self.opts
}

// Symbols returns the symbol table of the assembler.
func (self *Assembler) Symbols() *Symbols {
    return &self.repo
}

// Error generates a SyntaxError with msg.
func (self *Assembler) Error(msg string) *SyntaxError {
    return &SyntaxError {
        Pos    : -1,
        Row    : self.line.Row,
        Src    : self.line.Src,
        Reason : msg,
    }
}

// Rebase resets the origin to pc.
func (self *Assembler) Rebase(pc uintptr) *Assembler {
    self.repo.pc = pc
    return self
}

// Assemble assembles the assembly source and save the machine code to internal buffer.
func (self *Assembler) Assemble(src string) error {
    var err error
    var asm Parser
    var buf []*ParsedLine

    /* check if it is initialized */
    if self.arch == nil {
        panic("iasm: assembler was not initialized properly.")
    }

    /* parse the source */
    if buf, err = asm.init(self.arch).Parse(src); err != nil {
        return err
    }

    /* create a new program */
    p := self.arch.CreateProgram()
    defer p.Free()

    /* process every line */
    for _, self.line = range buf {
        switch self.cc++; self.line.Kind {
            case LineLabel       : if err = self.buildLabel(p, &self.line.Label)                  ; err != nil { return err }
            case LineCommand     : if err = self.buildCommand(p, &self.line.Command)              ; err != nil { return err }
            case LineInstruction : if err = self.arch.impl.Build(p, self, &self.line.Instruction) ; err != nil { return err }
            default              : panic("parser yields an invalid line kind")
        }
    }

    /* assemble the program */
    self.buf = p.Assemble(self.repo.pc)
    return nil
}
