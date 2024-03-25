package main

func throw(e error) {
	if e != nil {
		panic(e)
	}
}
func must[T any](t T, e error) T {
	throw(e)
	return t
}
