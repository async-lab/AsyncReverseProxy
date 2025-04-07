package goroutine

func MultiGo(num int, f func()) {
	for i := 0; i < num; i++ {
		go f()
	}
}
