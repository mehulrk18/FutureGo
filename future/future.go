package future

import (
	"errors"
	"fmt"

	condition "github.com/mehulrk18/FutureGo/condition"
)

// futureState enum type string
type futureState string
type fn func()

//
const (
	Pending              futureState = "PENDING"
	Running              futureState = "RUNNING"
	Cancelled            futureState = "CANCELLED"
	CancelledAndNotified futureState = "CANCELLED_AND_NOTIFIED"
	Finished             futureState = "FINISHED"
)

//GenFun generic Function struct to bind variables together.
type GenFun struct {
	Fun fn
}

// Futures Struct for asyncronous computation
type Futures struct {
	Condition     condition.Condition
	State         futureState
	Exception     error
	Result        int64
	Waiters       []*Waiter
	DoneCallbacks []fn
}

//Future is Future Channel
type Future struct {
	FutureChannel chan Futures
}

//InitFuture initiaizefuture
func (f *Future) InitFuture(read bool) {
	f.FutureChannel = make(chan Futures)
	var w []*Waiter
	var gf []fn
	cond := new(condition.Condition)
	cond.InitCondition(read)
	go func() {
		f.FutureChannel <- Futures{
			Condition:     *cond,
			State:         Pending,
			Exception:     nil,
			Result:        0,
			Waiters:       w,
			DoneCallbacks: gf,
		}
	}()
}

//InvokeCallBacks Execute callbacks
func (f *Future) InvokeCallBacks() {
	futureChan := <-f.FutureChannel
	for i := range futureChan.DoneCallbacks {
		futureChan.DoneCallbacks[i]()
	}
	go func() {
		f.FutureChannel <- futureChan
	}()
}

//GetState returns state of future
func (f *Future) GetState() string {
	futureChan := <-f.FutureChannel
	state := futureChan.State

	go func() {
		f.FutureChannel <- futureChan
	}()
	return string(state)
}

//Cancel the future if possible
func (f *Future) Cancel() bool {
	futureChan := <-f.FutureChannel
	response := true
	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == Running || futureChan.State == Finished {
			response = false
			futureChan.Condition.Release()
		} else if futureChan.State == Cancelled || futureChan.State == CancelledAndNotified {
			response = true
			futureChan.Condition.Release()
		} else {
			futureChan.State = Cancelled
			err = futureChan.Condition.NotifyAll()
		}
	}

	if err != nil {
		fmt.Println("Error while cancelling: ", err)
		response = false
	}
	go func() {
		f.FutureChannel <- futureChan
	}()
	return response
}

//IsCancelled returns true if future Cancelled.
func (f *Future) IsCancelled() bool {
	futureChan := <-f.FutureChannel
	response := false
	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == Cancelled || futureChan.State == CancelledAndNotified {
			response = true
		}
	}

	if err != nil {
		fmt.Println("Some error occured: ", err)
	}
	futureChan.Condition.Release()
	go func() {
		f.FutureChannel <- futureChan
	}()

	return response
}

//IsRunning checks whether the future is running or not.
func (f *Future) IsRunning() bool {
	futureChan := <-f.FutureChannel
	response := false

	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == Running {
			response = true
		}
	}

	if err != nil {
		fmt.Println("Some error occured: ", err)
	}
	futureChan.Condition.Release()
	go func() {
		f.FutureChannel <- futureChan
	}()

	return response
}

//IsDone checks whether the future is cancelled or finished executing.
func (f *Future) IsDone() bool {
	futureChan := <-f.FutureChannel
	response := false

	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == Finished || futureChan.State == Cancelled || futureChan.State == CancelledAndNotified {
			response = true
		}
	}

	if err != nil {
		fmt.Println("Some error occured: ", err)
	}
	futureChan.Condition.Release()
	go func() {
		f.FutureChannel <- futureChan
	}()

	return response
}

//GetResult Returns result or execption
func (f *Future) GetResult() (int64, error) {
	futureChan := <-f.FutureChannel
	result := futureChan.Result
	err := futureChan.Exception

	go func() {
		f.FutureChannel <- futureChan
	}()

	return result, err
}

//AddDoneCallBack adds callback to done call backs which will run when future finishes.
func (f *Future) AddDoneCallBack(fun fn) {
	futureChan := <-f.FutureChannel

	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State != Finished && futureChan.State != Cancelled && futureChan.State != CancelledAndNotified {
			futureChan.DoneCallbacks = append(futureChan.DoneCallbacks, fun)
		}
	} else {
		fun()
		if err != nil {
			fmt.Println("Some error callback: ", err)
		}
	}
	futureChan.Condition.Release()
	go func() {
		f.FutureChannel <- futureChan
	}()
}

//FinalResult gets result of callback that future represents.
func (f *Future) FinalResult(timeout uint32) (int64, error) {
	futureChan := <-f.FutureChannel
	result := int64(0)
	var err error

	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == CancelledAndNotified || futureChan.State == Cancelled {
			err = errors.New("Process was Cancelled")
		} else if futureChan.State == Finished {
			result, err = f.GetResult()
		} else {
			futureChan.Condition.Wait(timeout)
			if futureChan.State == CancelledAndNotified || futureChan.State == Cancelled {
				err = errors.New("Process was Cancelled")
			} else if futureChan.State == Finished {
				result, err = f.GetResult()
			} else {
				err = errors.New("Future Condition Timedout")
			}
		}
	}
	futureChan.Condition.Release()

	go func() {
		f.FutureChannel <- futureChan
	}()

	return result, err
}

//GetException returns exception of the future
func (f *Future) GetException(timeout uint32) error {
	futureChan := <-f.FutureChannel

	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == CancelledAndNotified || futureChan.State == Cancelled {
			err = errors.New("Process was Cancelled")
		} else if futureChan.State == Finished {
			err = futureChan.Exception
		} else {
			futureChan.Condition.Wait(timeout)
			if futureChan.State == CancelledAndNotified || futureChan.State == Cancelled {
				err = errors.New("Process was Cancelled")
			} else if futureChan.State == Finished {
				err = futureChan.Exception
			} else {
				err = errors.New("Future Condition Timedout")
			}
		}
	}
	futureChan.Condition.Release()
	go func() {
		f.FutureChannel <- futureChan
	}()

	return err
}

//SetRunningOrNotifyCancel mark future as running or process any cancel notification.
func (f *Future) SetRunningOrNotifyCancel() bool {
	futureChan := <-f.FutureChannel

	lock, err := futureChan.Condition.Aquire(true)
	response := true
	if lock && err == nil {
		if futureChan.State == Cancelled {
			futureChan.State = CancelledAndNotified
			// Add cancelled future to the waiters
			for _, w := range futureChan.Waiters {
				w.AddFutures(&futureChan)
			}

			response = false
		} else if futureChan.State == Pending {
			futureChan.State = Running
		} else {
			err := errors.New("Future in unexpected state: ")
			fmt.Println(err, futureChan.State)
			response = false
		}
	}
	futureChan.Condition.Release()
	go func() {
		f.FutureChannel <- futureChan
	}()

	return response
}

//SetResult mark future as running or process any cancel notification.
func (f *Future) SetResult(result int64) error {
	futureChan := <-f.FutureChannel

	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == Cancelled || futureChan.State == CancelledAndNotified || futureChan.State == Finished {
			err = fmt.Errorf("Invalid Result State: %s", futureChan.State)
			futureChan.Condition.Release()
		} else {
			futureChan.Result = result
			futureChan.State = Finished
			// Adding Completed futures
			for _, w := range futureChan.Waiters {
				w.AddFutures(&futureChan)
			}
			err = futureChan.Condition.NotifyAll()
		}
	}
	go func() {
		f.FutureChannel <- futureChan
	}()
	f.InvokeCallBacks()
	return err
}

//SetExecption sets expection for the Future
func (f *Future) SetExecption(exception error) error {
	futureChan := <-f.FutureChannel

	lock, err := futureChan.Condition.Aquire(true)

	if lock && err == nil {
		if futureChan.State == Cancelled || futureChan.State == CancelledAndNotified || futureChan.State == Finished {
			err = fmt.Errorf("Invalid Result State: %s", futureChan.State)

		} else {
			futureChan.Exception = exception
			futureChan.State = Finished
			// Adding futures with execptions
			for _, w := range futureChan.Waiters {
				w.AddFutures(&futureChan)
			}
			err = futureChan.Condition.NotifyAll()
		}
	}
	go func() {
		f.FutureChannel <- futureChan
	}()
	f.InvokeCallBacks()

	return err
}
