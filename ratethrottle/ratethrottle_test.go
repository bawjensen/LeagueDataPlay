package ratethrottle_test

import (
	"testing"
	"time"

	. "github.com/bawjensen/dataplay/ratethrottle"
	. "github.com/bawjensen/dataplay/utility"
)

func TestWait(t *testing.T) {
	start := time.Now()

	numOverflows := 3

	for i := 0; i < REQUEST_CAP * numOverflows; i++ {
		Wait()
	}

	expectedLower := time.Duration((numOverflows - 1) * REQUEST_PERIOD) * time.Second
	expectedUpper := time.Duration((numOverflows - 1) * REQUEST_PERIOD + 1) * time.Second

	// if elapsed := time.Now().Sub(start); elapsed < expectedSeconds {
	if elapsed := time.Now().Sub(start); !(elapsed > expectedLower && elapsed < expectedUpper) {
		t.Errorf("Outside expected time range (%v vs %v-%v)\n", elapsed, expectedLower, expectedUpper)
	}
}