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

type RateThrottle struct {
    buffer ringBuffer
    wait chan bool
}

var instance *RateThrottle

func init() {
    if instance != nil {
        fmt.Println("instance wasn't nil - How is this possible?")
    }

    instance = new(RateThrottle)
    (*instance) = NewRateThrottle()

    go func() {
        for {
            instance.wait <- true

            instance.buffer.ring[instance.buffer.curr] = time.Now()
            instance.buffer.increment()

            lastTime := instance.buffer.ring[instance.buffer.curr]
            if timeSince := time.Since(lastTime); (!lastTime.IsZero()) && (timeSince < instance.buffer.period) {
                // fmt.Printf("At %v, time is %v, sleeping for %v\n", instance.buffer.curr, instance.buffer.ring[instance.buffer.curr].Second(), instance.buffer.period - timeSince)
                time.Sleep(instance.buffer.period - timeSince)
            }
        }
    }()
}

func NewRateThrottle() (self RateThrottle) {
    self = RateThrottle{buffer: newRingBuffer(REQUEST_CAP, REQUEST_PERIOD * time.Second), wait: make(chan bool)}
    return
}

func Wait() {
    <-instance.wait
}