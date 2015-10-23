package ratethrottle

import (
	"fmt"
	"time"
	. "github.com/bawjensen/dataplay/utility"
)

// ------------------------------------- Global Variables ------------------------------------------

var instance *rateThrottle

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
	// rate 			int
	// ratePer 	time.Duration
	fillInterval	time.Duration
	// curr 			int
	// last_check 		time.Time
	wait 			chan signal
}

// func newRateThrottle() (self rateThrottle) {
// 	size := int(REQUEST_CAP / RATE_THROTTLE_GRANULARITY)
// 	time := ((REQUEST_PERIOD + RATE_THROTTLE_BUFFER) * 1000 * time.Millisecond) / RATE_THROTTLE_GRANULARITY
// 	fmt.Printf("Initializing rate throttler with size %d and time %v\n", size, time)
// 	self = rateThrottle{buffer: newRingBuffer(int(size), time), wait: make(chan signal)}
// 	return
// }

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
	<-instance.wait
}

// ------------------------------------- Package Init Logic ----------------------------------------

func init() {
	if instance != nil {
		fmt.Println("instance wasn't nil - How is this possible?")
	}

	instance = new(rateThrottle) 	// Allocate an actual space for non-nil rateThrottle
	(*instance) = newRateThrottle() // Populate that

	go func() {
		for _ = range time.Tick(instance.fillInterval) {
			instance.wait <- signal{}
		}
	}()

	// go func() {
	// 	for {
	// 		instance.wait <- true

	// 		instance.buffer.setCurrent(time.Now())
	// 		instance.buffer.increment()

	// 		lastTime := instance.buffer.current()
	// 		if timeSince := time.Since(lastTime); (!lastTime.IsZero()) && (timeSince < instance.buffer.period) {
	// 			// fmt.Printf("At %v, time is %v, sleeping for %v\n", instance.buffer.curr, instance.buffer.ring[instance.buffer.curr].Second(), instance.buffer.period - timeSince)
	// 			time.Sleep(instance.buffer.period - timeSince)
	// 		}
	// 	}
	// }()
}