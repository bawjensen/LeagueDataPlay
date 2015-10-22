package api

import (
    "fmt"
    "time"
)

// -------------------------------------- Global variables -----------------------------------------

var eventReportChan chan byte

// -------------------------------------- Reporter logic -------------------------------------------

const (
    REQUEST_SEND_EVENT = iota
    REQUEST_SUCCESS_EVENT
    TIMEOUT_EVENT
    RESET_EVENT
    SERV_RATE_LIMIT_EVENT
    USER_RATE_LIMIT_EVENT
    UNKNOWN_ERROR_EVENT
    SERVER_ERROR_EVENT

    NUM_ERRORS // As long as it's at the end, will correctly reflect the number of "enums" in this const block
)

func init() {
    // Set up event reporting chan, for nice report outputs
    eventReportChan = make(chan byte)

    // Set up event listener and reporter
    var events [NUM_ERRORS]int

    go func() {
        var eventType byte

        for {
            eventType = <-eventReportChan
            events[eventType]++
        }
    }()

    go func() {
        for _ = range time.Tick(200 * time.Millisecond) {
            fmt.Printf("\rAt %d (%d) req's, %d (%d) rate-lim, %d serv-err, %d t/o, %d resets, %d other errors",
                events[REQUEST_SUCCESS_EVENT],
                events[REQUEST_SEND_EVENT],
                events[USER_RATE_LIMIT_EVENT],
                events[SERV_RATE_LIMIT_EVENT],
                events[SERVER_ERROR_EVENT],
                events[TIMEOUT_EVENT],
                events[RESET_EVENT],
                events[UNKNOWN_ERROR_EVENT])
        }
    }()
}
