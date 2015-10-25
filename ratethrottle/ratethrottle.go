package ratethrottle

import (
	"fmt"
	"log"
	"time"
	. "github.com/bawjensen/dataplay/utility"
)

// ------------------------------------- Global Variables ------------------------------------------

var instance *rateThrottle
var sleepTime time.Duration

// ------------------------------------- Ring Buffer Logic -----------------------------------------

type ringBuffer struct {
	ring []time.Time
	size int
	period time.Duration
	curr int
}

func newRingBuffer(size int, period time.Duration) (self ringBuffer) {
	self = ringBuffer{ring: make([]time.Time, size, size), curr: 0, size: size, period: period}

	return self
}

func (self *ringBuffer) increment() {
	self.curr = (self.curr + 1) % self.size
}

func (self *ringBuffer) current() time.Time {
	return self.ring[self.curr]
}

func (self *ringBuffer) setCurrent(newTime time.Time) {
	self.ring[self.curr] = newTime
}

// ------------------------------------- Rate Throttle Logic ---------------------------------------

type signal struct{}

type rateThrottle struct {
	fillInterval	time.Duration
	wait 			chan signal
}

func newRateThrottle() (self rateThrottle) {
	fillInterval := (REQUEST_PERIOD + RATE_THROTTLE_BUFFER) / REQUEST_CAP
	bufferSize := int(REQUEST_CAP / RATE_THROTTLE_GRANULARITY)

	fmt.Printf("Initializing rate throttler with interval %v and buffer %d\n", fillInterval, bufferSize)

	self = rateThrottle{
		fillInterval: 	fillInterval,
		wait: 			make(chan signal, bufferSize),
	}

	return
}

// ------------------------------------- Package behavior Logic ------------------------------------

func Wait() {
	if (sleepTime != 0) {
		log.Println("Sleeping all requests for", sleepTime)
		time.Sleep(sleepTime)
		sleepTime = 0
	}
	<-instance.wait
}

func Sleep(dur time.Duration) {
	sleepTime = dur
}

// ------------------------------------- Package Init Logic ----------------------------------------

func init() {
	if instance != nil {
		log.Fatal("instance wasn't nil - How is this possible?")
	}

	instance = new(rateThrottle) 	// Allocate an actual space for non-nil rateThrottle
	(*instance) = newRateThrottle() // Populate that

	go func() {
		for _ = range time.Tick(instance.fillInterval) {
			instance.wait <- signal{}
		}
	}()
}