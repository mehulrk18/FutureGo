package future

//Submit function
func Submit(fun Fn) *Future {
	var f *Future
	f = new(Future)
	*f = InitFuture(true)

	WorkItemInit(f, fun)

	return f
}
