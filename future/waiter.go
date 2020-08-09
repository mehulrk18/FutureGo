package future

import (
	condition "github.com/mehulrk18/FutureGo/condition"
)

//Waiter provides events that waits and completed.
type Waiter struct {
	Event           condition.Event
	FinishedFutures []*Futures
}

//InitWaiter Creates instance of waitier.
func InitWaiter() *Waiter {
	var w *Waiter
	var futureList []*Futures
	w.Event = condition.Event{}
	w.Event.InitEvent()
	w.FinishedFutures = futureList

	return w
}

// AddFutures to waiter with results, execptions or cancelled..
func (w *Waiter) AddFutures(f *Futures) {

	w.FinishedFutures = append(w.FinishedFutures, f)
}
