# `IASM` -- Interactive Assembler for Go.

Dual-purpose assembly engine written in pure Golang.

The `x86_64` package was ported from a Python module [`PeachPy`](https://github.com/Maratyszcza/PeachPy), with some adaption to the Go language features.

**Currently, IASM only supports x86_64, because it's the only architecture I am very familiar with.**

### It can be used as a dynamic assembler

This can be used to implement:

* JIT engine
* Compiler backend
* And more !

See `x86_64/program_test.go` for more info.

### It can also be used as a static assembler

For `macOS`:

```bash
git clone https://github.com/chenzhuoyu/iasm
cd iasm
go build ./cmd/iasm
./iasm -h
./iasm -f macho -D __Darwin__ -o helloworld example/helloworld.s
./helloworld 
```

### It also contains an interactive REPL shell

Just run IASM without any arguments.

```bash
./iasm
```
