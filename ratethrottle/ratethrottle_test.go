package ratethrottle_test

import (
	// "fmt"
	"testing"
	"time"

	. "github.com/bawjensen/dataplay/ratethrottle"
	. "github.com/bawjensen/dataplay/utility"
)

func TestWait(t *testing.T) {
	start := time.Now()

	numRequestCapHits := 5

	for i := 0; i < REQUEST_CAP * numRequestCapHits; i++ {
		Wait()
	}

	expectedLower := time.Duration(numRequestCapHits) * REQUEST_PERIOD * time.Second
	expectedUpper := time.Duration(numRequestCapHits + 1) * REQUEST_PERIOD * time.Second

	// if elapsed := time.Now().Sub(start); elapsed < expectedSeconds {
	if elapsed := time.Now().Sub(start); !(elapsed > expectedLower && elapsed < expectedUpper) {
		t.Errorf("Outside expected time range (%v vs %v-%v)\n", elapsed, expectedLower, expectedUpper)
	}
}