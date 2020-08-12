package future

import (
	"errors"
	"fmt"

	condition "github.com/mehulrk18/FutureGo/condition"
)

// futureState enum type string
type futureState string

//Fn Function for future to execute.
type Fn func() int64 //x, y int64

//
const (
	Pending              futureState = "PENDING"
	Running              futureState = "RUNNING"
	Cancelled            futureState = "CANCELLED"
	CancelledAndNotified futureState = "CANCELLED_AND_NOTIFIED"
	Finished             futureState = "FINISHED"
)

// Future Struct for asyncronous computation
type Future struct {
	Condition     condition.Condition
	State         futureState
	Exception     error
	Result        int64
	Waiters       []*Waiter
	DoneCallbacks []Fn
}

//InitFuture initiaizefuture
func InitFuture(read bool) Future {
	var future Future
	var w []*Waiter
	var gf []Fn
	cond := new(condition.Condition)
	cond.InitCondition(read)
	future = Future{
		Condition:     *cond,
		State:         Pending,
		Exception:     nil,
		Result:        0,
		Waiters:       w,
		DoneCallbacks: gf,
	}
	return future
}

//InvokeCallBacks Execute callbacks
func (f *Future) InvokeCallBacks() {
	for i := range f.DoneCallbacks {
		f.DoneCallbacks[i]() // FIX THIS
	}
}

//GetState returns state of future
func (f *Future) GetState() string {
	state := f.State
	return string(state)
}

//Cancel the future if possible
func (f *Future) Cancel() bool {
	response := true
	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == Finished {
			response = false
			f.Condition.Release()
		} else if f.State == Cancelled || f.State == CancelledAndNotified {
			response = true
			f.Condition.Release()
		} else {
			f.State = Cancelled
			err = f.Condition.NotifyAll()
		}
	}

	if err != nil {
		fmt.Println("Error while cancelling: ", err)
		response = false
	}
	return response
}

//IsCancelled returns true if future Cancelled.
func (f *Future) IsCancelled() bool {
	response := false
	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == Cancelled || f.State == CancelledAndNotified {
			response = true
		}
	}

	if err != nil {
		fmt.Println("Some error occured: ", err)
	}
	f.Condition.Release()

	return response
}

//IsRunning checks whether the future is running or not.
func (f *Future) IsRunning() bool {
	response := false

	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == Running {
			response = true
		}
	}

	if err != nil {
		fmt.Println("Some error occured: ", err)
	}
	f.Condition.Release()

	return response
}

//IsDone checks whether the future is cancelled or finished executing.
func (f *Future) IsDone() bool {
	response := false

	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == Finished || f.State == Cancelled || f.State == CancelledAndNotified {
			response = true
		}
	}

	if err != nil {
		fmt.Println("Some error occured: ", err)
	}
	f.Condition.Release()

	return response
}

//GetResult Returns result or execption
func (f *Future) GetResult() (int64, error) {
	result := f.Result //f.Result
	err := f.Exception //f.Exception

	if err != nil {
		return 0, err
	}
	return result, nil
}

//AddDoneCallBack adds callback to done call backs which will run when future finishes.
func (f *Future) AddDoneCallBack(fun Fn) {
	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State != Finished && f.State != Cancelled && f.State != CancelledAndNotified {
			f.DoneCallbacks = append(f.DoneCallbacks, fun)
		}
	} else {
		fun()
		if err != nil {
			fmt.Println("Some error callback: ", err)
		}
	}
	f.Condition.Release()
}

//FinalResult gets result of callback that future represents.
func (f *Future) FinalResult(timeout uint32) (int64, error) {
	result := int64(0)
	var err error

	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == CancelledAndNotified || f.State == Cancelled {
			err = errors.New("Process was Cancelled")
		} else if f.State == Finished {
			result, err = f.GetResult()
		} else {
			f.Condition.Wait(timeout)
			if f.State == CancelledAndNotified || f.State == Cancelled {
				err = errors.New("Process was Cancelled")
			} else if f.State == Finished {
				result, err = f.GetResult()
			} else {
				err = errors.New("Future Condition Timedout")
			}
		}
	}
	f.Condition.Release()
	return result, err
}

//GetException returns exception of the future
func (f *Future) GetException(timeout uint32) error {
	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == CancelledAndNotified || f.State == Cancelled {
			err = errors.New("Process was Cancelled")
		} else if f.State == Finished {
			err = f.Exception
		} else {
			f.Condition.Wait(timeout)
			if f.State == CancelledAndNotified || f.State == Cancelled {
				err = errors.New("Process was Cancelled")
			} else if f.State == Finished {
				err = f.Exception
			} else {
				err = errors.New("Future Condition Timedout")
			}
		}
	}
	f.Condition.Release()
	return err
}

//SetRunningOrNotifyCancel mark future as running or process any cancel notification.
func (f *Future) SetRunningOrNotifyCancel() bool {
	lock, err := f.Condition.Aquire(true)
	response := true
	if lock && err == nil {
		if f.State == Cancelled {
			f.State = CancelledAndNotified
			// Add cancelled future to the waiters
			for _, w := range f.Waiters {
				w.AddFutures(f)
			}

			response = false
		} else if f.State == Pending {
			f.State = Running
		} else {
			err := errors.New("Future in unexpected state: ")
			fmt.Println(err, f.State)
			response = false
		}
	}
	f.Condition.Release()
	return response
}

//SetResult mark future as running or process any cancel notification.
func (f *Future) SetResult(result int64) error {
	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == Cancelled || f.State == CancelledAndNotified || f.State == Finished {
			err = fmt.Errorf("Cannot Set Result process was %s ", f.State)
			f.Condition.Release()
		} else {
			f.Result = result
			f.State = Finished
			// Adding Completed futures
			for _, w := range f.Waiters {
				w.AddFutures(f)
			}
			err = f.Condition.NotifyAll()
		}
	}
	if err == nil {
		f.InvokeCallBacks()
	}
	return err
}

//SetExecption sets expection for the Future
func (f *Future) SetExecption(exception error) error {
	lock, err := f.Condition.Aquire(true)

	if lock && err == nil {
		if f.State == Cancelled || f.State == CancelledAndNotified || f.State == Finished {
			err = fmt.Errorf("Invalid Result State to set execption: %s", f.State)

		} else {
			f.Exception = exception
			f.State = Finished
			// Adding futures with execptions
			for _, w := range f.Waiters {
				w.AddFutures(f)
			}
			err = f.Condition.NotifyAll()
		}
	}
	if err == nil {
		f.InvokeCallBacks()
	}
	return err
}
