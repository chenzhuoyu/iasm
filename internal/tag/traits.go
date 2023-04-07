package tag

type (
    Tag    struct    {}
    Sealed interface { Sealed(Tag) }
)

type Disposable interface {
    // Free returns the object into pool.
    // Any operation performed after Free is undefined behavior.
    Free()
}

type Verifiable interface {
    EnsureValid()
}
