package misc

func Ref[T any](t T) *T {
	return &t
}
