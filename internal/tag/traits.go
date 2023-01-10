package tag

type (
    Tag    struct{}
    Sealed interface{ Sealed(Tag) }
)

type Disposable interface {
    Free()
}

type Verifiable interface {
    EnsureValid()
}
