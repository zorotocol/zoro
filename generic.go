package oracle

func Throw(e error) {
	if e != nil {
		panic(e)
	}
}
func Must[T any](t T, e error) T {
	Throw(e)
	return t
}
