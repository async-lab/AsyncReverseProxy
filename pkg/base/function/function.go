package function

type Supplier[T any] func() T
type Consumer[T any] func(T)
type Predicate[T any] func(T) bool
