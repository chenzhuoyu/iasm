package repl

import (
    `errors`

    `github.com/chenzhuoyu/iasm/x86_64`
)

type _IASMArchSpecific struct {
    ps x86_64.Parser
}

func (self *_IASMArchSpecific) doasm(addr uintptr, line string) ([]byte, error) {
    var err error
    var asm x86_64.Assembler
    var ast *x86_64.ParsedLine

    /* parse the line */
    if ast, err = self.ps.Feed(line); err != nil {
        return nil, err
    }

    /* interactive shell does not support labels */
    if ast.Kind == x86_64.LineLabel {
        return nil, errors.New("interactive shell does not support labels")
    }

    /* assemble the line */
    if err = asm.WithBase(addr).Assemble(line); err != nil {
        return nil, err
    } else {
        return asm.Code(), nil
    }
}
