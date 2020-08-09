package condition

import (
	"fmt"
)

//EventMember for implementing EventMember
type EventMember struct {
	Cond  Condition
	Flag  bool
	isSet bool
}

// Event channel for event
type Event struct {
	EventChannel chan EventMember
}

// InitEvent initialize Event
func (e *Event) InitEvent() {
	e.EventChannel = make(chan EventMember)
	cond := new(Condition)
	cond.InitCondition(false)
	go func() {
		e.EventChannel <- EventMember{
			Cond: *cond,
			Flag: false,
		}
	}()
}

//resetInternalLocks called at fork.
func (e *Event) resetInternalLocks() {
	eventChan := <-e.EventChannel

	eventChan.Cond.InitCondition(false)

	go func() {
		e.EventChannel <- eventChan
	}()
}

//IsSet Return true if and only if the internal flag is true.
func (e *Event) IsSet() bool {
	eventChan := <-e.EventChannel
	eventChan.isSet = eventChan.Flag

	go func() {
		e.EventChannel <- eventChan
	}()

	return eventChan.isSet
}

// Set sets internal flag to true.
func (e *Event) Set() {
	eventChan := <-e.EventChannel

	aquired, err := eventChan.Cond.Aquire(true)

	if aquired && err != nil {
		eventChan.Flag = true
		err = eventChan.Cond.NotifyAll()
	}
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		e.EventChannel <- eventChan
	}()
}

//Clear sets internal flag to false.
func (e *Event) Clear() {
	eventChan := <-e.EventChannel

	aquired, err := eventChan.Cond.Aquire(true)

	if aquired && err != nil {
		eventChan.Flag = false
		err = eventChan.Cond.Release()
	}
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		e.EventChannel <- eventChan
	}()
}

//Wait blocks until the internal flag is True
func (e *Event) Wait() bool {
	eventChan := <-e.EventChannel

	aquired, err := eventChan.Cond.Aquire(true)

	if aquired && err != nil {
		signaled := eventChan.Flag
		if !signaled {
			signaled = eventChan.Cond.Wait(0)
		}
		return signaled
	}
	return false
}
