package api

import (
	. "github.com/bawjensen/dataplay/utility"
)

// ----------------------------------------- Globals -----------------------------------------------

var simulRequestLimiter chan signal

// -------------------------------------- Limiter logic --------------------------------------------

type signal struct{}

func init() {
	// Initialize simultaneous request limiter
	simulRequestLimiter = make(chan signal, MAX_SIMUL_REQUESTS)

	// Set up simultaneous request limiter with full allotment
	for i := 0; i < MAX_SIMUL_REQUESTS; i++ {
		simulRequestLimiter <- signal{}
	}
}
