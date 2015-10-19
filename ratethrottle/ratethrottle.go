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
    fmt.Println("Initializing rate throttler")

    instance = new(RateThrottle)
    (*instance) = newRateThrottle()

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
    // Increase the time of the rate throttle by 1.1, to account for small issues with rate limiting
    self = RateThrottle{buffer: newRingBuffer(REQUEST_CAP / RATE_THROTTLE_GRANULARITY, ((REQUEST_PERIOD + RATE_THROTTLE_BUFFER) * time.Second) / RATE_THROTTLE_GRANULARITY), wait: make(chan bool)}
    return
}

func Wait() {
    <-instance.wait
}