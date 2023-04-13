package aarch64

import (
    `github.com/chenzhuoyu/iasm/asm`
)

type _ParserImpl struct {
    lex *asm.Tokenizer
}

func mkparser(lex *asm.Tokenizer) *_ParserImpl {
    return &_ParserImpl { lex }
}

func (self *_ParserImpl) parse(ins *asm.ParsedInstruction) {
    // TODO: this
    panic("aarch64: _ParserImpl.parse(): not implemented")
}
