package condition

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ThreadCondition Struct is Go Condition Varaible.
type ThreadCondition struct {
	Lock    sync.RWMutex
	Read    bool //true: for reading, false: for writing
	Locked  bool //true: Locked, false: unlocked
	Block   bool //true: block thread
	Waiters sync.WaitGroup
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
	go func() {
		c.ConditionChannel <- ThreadCondition{
			Lock:    sync.RWMutex{},
			Locked:  false,
			Read:    reading,
			Block:   false,
			Waiters: sync.WaitGroup{},
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
	var err error
	if blocking {

		if channelVal.Locked && !channelVal.Read {
			blocking = false
			err = errors.New("Lock Already Aquired")

		} else {
			if channelVal.Block { // Channel is being used by someone
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
		}
	} else {
		if channelVal.Block {
			blocking = false
		} else {
			blocking = true
		}
		err = nil
	}

	go func() {
		c.ConditionChannel <- channelVal
	}()

	if err != nil {
		return false, err
	}

	return blocking, nil
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
func (c *Condition) IsOwned() bool {
	blocked, _ := c.Aquire(false)

	if blocked {
		c.Release()
		return false
	}
	return true
}

//Wait waits until Notified or Timeout Occurs
func (c *Condition) Wait(timer uint32) {
	channelVal := <-c.ConditionChannel
	chanwait := make(chan bool)
	var timedout bool
	go func() {
		channelVal.Timeout(timer, chanwait)
		timedout = <-chanwait
	}()
	if timedout {
		fmt.Println("Waiting Timed out")
	} else {
		fmt.Println("Task Completed")
	}
	go func() {
		c.ConditionChannel <- channelVal
	}()

}

//Notify n number of waitng tasks.
func (c *Condition) Notify(n uint32) error {
	if !c.IsOwned() {
		return errors.New("cannot notify on un-acquired lock")
	}
	var err error
	channelVal := <-c.ConditionChannel
	if channelVal.Total == 0 {
		err = errors.New("No task to notify")
	}
	err = nil
	var i uint32

	for i = 0; i < n; i++ {
		if channelVal.Total == 0 {
			err = errors.New("No more tasks left to notify")
		}
		channelVal.Waiters.Done()
		channelVal.Total--
	}

	go func() {
		c.ConditionChannel <- channelVal
	}()
	return err
}

//NotifyAll n number of waitng tasks.
func (c *Condition) NotifyAll() error {
	channelVal := <-c.ConditionChannel
	err := c.Notify(channelVal.Total)
	go func() {
		c.ConditionChannel <- channelVal
	}()
	return err
}
