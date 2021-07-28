# `IASM` -- Interactive Assembler for Go.

Dual-purpose assembly engine written in pure Golang.

The `x86_64` package was ported from a Python module [`PeachPy`](https://github.com/Maratyszcza/PeachPy), with some adaption to the Go language features.

**Currently, IASM only supports x86_64, because it's the only architecture I am very familiar with.**

### It can be used as a dynamic assembler

This can be used to implement:

* JIT engine
* Compiler backend
* And more !

### It can also be used as a static assembler
