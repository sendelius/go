package users

type BeforeCreateHook[T Accessor] func(T) error

type AfterCreateHook[T Accessor] func(T) error

type Hooks[T Accessor] struct {
	BeforeCreate []BeforeCreateHook[T]
	AfterCreate  []AfterCreateHook[T]
}
