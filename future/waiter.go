package future

import (
	condition "github.com/mehulrk18/FutureGo/condition"
)

//Waiter provides events that waits and completed.
type Waiter struct {
	Event           condition.Event
	FinishedFutures []*Future
}

//InitWaiter Creates instance of waitier.
func InitWaiter() *Waiter {
	var w *Waiter
	var futureList []*Future
	w.Event = condition.Event{}
	w.Event.InitEvent()
	w.FinishedFutures = futureList

	return w
}

// AddFutures to waiter with results, execptions or cancelled..
func (w *Waiter) AddFutures(f *Future) {

	w.FinishedFutures = append(w.FinishedFutures, f)
}
