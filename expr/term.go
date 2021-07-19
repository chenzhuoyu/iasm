package expr

// Term represents a value that can be Evaluate()'d into an integer.
type Term interface {
    Evaluate() int64
}
