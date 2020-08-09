package future

import (
	"errors"
	"fmt"
)

//TestFuture testing Future
func TestFuture(f fn) Future {
	fut := new(Future)
	fut.InitFuture(true)

	// True: read lock, False: write lock

	for {
		fmt.Println(" state: ", fut.GetState())
		fut.AddDoneCallBack(f)

		// fut.Cancel()

		fmt.Println("Running: ", fut.SetRunningOrNotifyCancel())

		fmt.Println(" state: ", fut.GetState())

		fut.SetResult(5)
		fut.SetExecption(errors.New("Some Exception"))

		fut.InvokeCallBacks()
		fut.GetResult()

		if fut.GetState() == string(Finished) || fut.GetState() == string(Cancelled) || fut.GetState() == string(CancelledAndNotified) {
			fmt.Println(" state: ", fut.GetState())
			break
		}

	}

	return *fut

}
