package future

import "fmt"

//WorkItem struct
type WorkItem struct {
	Future   *Future
	Function Fn
	// x, y		int64
}

// WorkItemInit Initialize work Item.
func WorkItemInit(f *Future, fun Fn) WorkItem {
	w := WorkItem{
		Future:   f,
		Function: fun,
		// x:				x,
		// y:				y,
	}

	w.run()

	return w
}

// run cancel param to cancel the current execution.
func (w *WorkItem) run() {
	// fmt.Println("Current State", w.Future.GetState())
	if !w.Future.SetRunningOrNotifyCancel() {
		return
	}
	// w.Future.Cancel()

	result := w.Function()
	// fmt.Println("Current State", w.Future.GetState())
	err := w.Future.SetResult(result)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("Current  State", w.Future.GetState())

}
