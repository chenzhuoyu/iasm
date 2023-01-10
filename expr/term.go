package expr

import (
    `github.com/chenzhuoyu/iasm/internal/tag`
)

// Term represents a value that can Evaluate() into an integer.
type Term interface {
    tag.Disposable
    Evaluate() (int64, error)
}
