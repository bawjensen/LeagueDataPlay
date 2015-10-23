package api

import (
	"fmt"
	"time"
)

// -------------------------------------- Global variables -----------------------------------------

var eventReportChan chan byte
const (
	REPORT_INTERVAL = 500 * time.Millisecond
)

// -------------------------------------- Reporter logic -------------------------------------------

const (
	REQUEST_SEND_EVENT = iota // Sent a request
	REQUEST_SUCCESS_EVENT 	// Request went just fine
	REQUEST_AVOIDED_EVENT 	// Any time a request might have been sent, but didn't have to be

	TIMEOUT_EVENT 			// Connection timeout event, caused by (?)
	RESET_EVENT 			// Connection reset event, caused by (?)

	SERV_RATE_LIMIT_EVENT	// 429 caused by their infrastructure
	USER_RATE_LIMIT_EVENT	// 429 caused by me and associated with given API key

	SERVER_ERROR_EVENT 		// Error on their end, a 5XX
	UNKNOWN_ERROR_EVENT		// Error that's not quite known

	NUM_ERRORS 				// As long as it's at the end, will correctly reflect the number of "enums" in this const block
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
		for _ = range time.Tick(REPORT_INTERVAL) {
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
