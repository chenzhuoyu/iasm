package expr

// Term represents a value that can Evaluate() into an integer.
type Term interface {
    Evaluate() (int64, error)
}
