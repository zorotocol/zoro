package misc

func Throw(e error) {
	if e != nil {
		panic(e)
	}
}
func Must[T any](t T, e error) T {
	Throw(e)
	return t
}

func ErrChan(fn func() error) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- fn()
	}()
	return ch
}
