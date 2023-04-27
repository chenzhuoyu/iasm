package expr

import (
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Term represents a value that can Evaluate() into an integer.
type Term interface {
    tag.Disposable
    Evaluate() (int64, error)
}

type _CurrentPos struct {
    r Repository
}

func mkcurpos(r Repository) (tr _CurrentPos) {
    tr.r = r
    return
}

func (self _CurrentPos) Free() {}
func (self _CurrentPos) Evaluate() (int64, error) { return self.r.Pos(), nil }
