package ratethrottle

import (
    "fmt"
    "time"
    . "github.com/bawjensen/dataplay/utility"
)

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

type RateThrottle struct {
    buffer ringBuffer
    wait chan bool
}

var instance *RateThrottle

func init() {
    if instance != nil {
        fmt.Println("instance wasn't nil - How is this possible?")
    }

    instance = new(RateThrottle) // Allocate an actual space for non-nil RateThrottle
    (*instance) = newRateThrottle() // Populate that

    go func() {
        for {
            instance.wait <- true

            instance.buffer.setCurrent(time.Now())
            instance.buffer.increment()

            lastTime := instance.buffer.current()
            if timeSince := time.Since(lastTime); (!lastTime.IsZero()) && (timeSince < instance.buffer.period) {
                // fmt.Printf("At %v, time is %v, sleeping for %v\n", instance.buffer.curr, instance.buffer.ring[instance.buffer.curr].Second(), instance.buffer.period - timeSince)
                time.Sleep(instance.buffer.period - timeSince)
            }
        }
    }()
}

func newRateThrottle() (self RateThrottle) {
    size := int(REQUEST_CAP / RATE_THROTTLE_GRANULARITY)
    time := ((REQUEST_PERIOD + RATE_THROTTLE_BUFFER) * 1000 * time.Millisecond) / RATE_THROTTLE_GRANULARITY
    fmt.Printf("Initializing rate throttler with size %d and time %v\n", size, time)
    self = RateThrottle{buffer: newRingBuffer(int(size), time), wait: make(chan bool)}
    return
}

func Wait() {
    <-instance.wait
}