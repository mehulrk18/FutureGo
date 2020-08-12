package condition

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ThreadCondition Struct is Go Condition Varaible.
type ThreadCondition struct {
	Lock    *sync.RWMutex
	Read    bool //true: for reading, false: for writing
	Locked  bool //true: Locked, false: unlocked
	Block   bool //true: block thread
	Waiters *sync.WaitGroup
	Total   uint32 // count number of tasks
}

//Timeout tell wait time has timedout or not.
func (tc *ThreadCondition) Timeout(timer uint32, chanwait chan bool) {
	chanwait <- tc.Locked
	if tc.Locked {
		tc.Waiters.Wait()
		time.Sleep(time.Duration(timer) * time.Second)
		chanwait <- tc.Locked
	}
}

// Condition is Go Condition channel.
type Condition struct {
	ConditionChannel chan ThreadCondition
}

// InitCondition Initializes new ThreadCondition variable
func (c *Condition) InitCondition(reading bool) {
	c.ConditionChannel = make(chan ThreadCondition)
	rw := new(sync.RWMutex)
	wg := new(sync.WaitGroup)
	go func() {
		c.ConditionChannel <- ThreadCondition{
			Lock:    rw,
			Locked:  false,
			Read:    reading,
			Block:   false,
			Waiters: wg,
			Total:   0,
		}
	}()
}

// AquireRestore aquires lock.
func (c *Condition) AquireRestore() error {
	_, err := c.Aquire(true)
	return err
}

// ReleaseSave releases lock.
func (c *Condition) ReleaseSave() error {
	return c.Release()
}

// Aquire to allow blocking
func (c *Condition) Aquire(blocking bool) (bool, error) {
	channelVal := <-c.ConditionChannel

	var aquired bool
	var err error

	if blocking {

		if channelVal.Locked && !channelVal.Read {
			aquired = false
			err = errors.New("Lock Already Aquired")

		} else {
			if channelVal.Block && !channelVal.Read { // Channel is being used by someone
				channelVal.Waiters.Wait()
			}
			channelVal.Locked = true
			channelVal.Block = true
			if channelVal.Read {
				channelVal.Lock.RLock()
			} else {
				channelVal.Lock.Lock()
			}
			channelVal.Waiters.Add(1)
			channelVal.Total++
			aquired = true
		}
	} else {
		if channelVal.Block {
			aquired = false
		} else {
			aquired = true
		}
		err = nil
	}

	go func() {
		c.ConditionChannel <- channelVal
	}()

	return aquired, err
}

//Release to release lock
func (c *Condition) Release() error {
	channelValue := <-c.ConditionChannel
	var err error
	if !channelValue.Locked {
		err = errors.New("Cannot Release an un-aquired lock")
	} else {
		if channelValue.Read {
			channelValue.Lock.RUnlock()
		} else {
			channelValue.Lock.Unlock()
		}
		channelValue.Block = false
		channelValue.Locked = false
		channelValue.Waiters.Done()
		channelValue.Total--
		err = nil
	}

	go func() {
		c.ConditionChannel <- channelValue
	}()

	return err
}

// IsOwned returns true if lock is owned by current thread.
// func (c *Condition) IsOwned() bool {
// 	channelVal := <-c.ConditionChannel
// 	blocked, _ := c.Aquire(false)

// 	if blocked {
// 		c.Release()
// 		return false
// 	}
// 	go func() {
// 		c.ConditionChannel <- channelVal
// 	}()
// 	return true
// }

//Wait waits until Notified or Timeout Occurs
func (c *Condition) Wait(timer uint32) bool {
	if timer == 0 {
		c.Aquire(true)
		return true
	}
	channelVal := <-c.ConditionChannel
	chanwait := make(chan bool)
	var timedout, gotIt bool

	go func() {
		c.Aquire(true)
		gotIt = true
		channelVal.Timeout(timer, chanwait)
		timedout = <-chanwait
	}()
	if timedout {
		fmt.Println("Waiting Timed out")
		gotIt = false
	} else {
		fmt.Println("Task Completed")
	}
	go func() {
		c.ConditionChannel <- channelVal
	}()

	return gotIt
}

//Notify n number of waitng tasks in the thread.
func (tc *ThreadCondition) Notify(n uint32) error {
	var err error
	if tc.Block {
		if !tc.Locked {
			err = errors.New("cannot notify on un-acquired lock")
		} else {
			if tc.Read {
				tc.Lock.RUnlock()
			} else {
				tc.Lock.Unlock()
			}
			tc.Block = false
			tc.Locked = false
			tc.Waiters.Done()
			tc.Total--
			err = nil
		}
	} else {
		if tc.Total == 0 {
			err = errors.New("No task to notify")
		}
		err = nil
		var i uint32

		for i = 0; i < n; i++ {
			if tc.Total == 0 {
				err = errors.New("No more tasks left to notify")
			}
			tc.Waiters.Done()
			tc.Total--
		}
	}
	return err
}

//NotifyAll n number of waitng tasks of the channel.
func (c *Condition) NotifyAll() error {
	channelVal := <-c.ConditionChannel
	n := channelVal.Total
	err := channelVal.Notify(n)
	go func() {
		c.ConditionChannel <- channelVal
	}()
	return err
}
