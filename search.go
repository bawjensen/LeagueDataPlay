package main

import (
	"flag"
	"fmt"
	"log"
	// "math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	// "sync"
	"time"

	. "github.com/bawjensen/dataplay/api"
	. "github.com/bawjensen/dataplay/utility"
	// . "github.com/bawjensen/dataplay/constants"
)

// ------------------------------------ Global variables -------------------------------------------

var simulRequestLimiter chan bool

// ------------------------------------ Search logic -----------------------------------------------


func MakeIterator(slice []interface{}) chan interface{} {
	ch := make(chan interface{})

	go func() {
		for i := 0; i < len(slice); i++ { // Important to iterate by index for some reason - range doesn't work
			ch <- slice[i]
		}

		close(ch)
	}()

	return ch
}


func partitionByNum(input []interface{}, num int) [][]interface{} {
	inputSize := len(input)
	sliceSize := inputSize / num
	numOverloadedSlices := inputSize % num // number of slices with an extra element

	slices := make([][]interface{}, num, num)

	iterator := MakeIterator(input)

	for i := 0; i < num; i++ {
		numInSlice := sliceSize
		if i < numOverloadedSlices {
			numInSlice += 1
		}

		slices[i] = make([]interface{}, numInSlice, numInSlice)

		for j := 0; j < numInSlice; j++ {
			slices[i][j] = <-iterator
		}
	}

	return slices
}


func createSliceHandler(mapper func(interface{}, []*IntSet) (*IntSet, *IntSet), in chan []interface{}, out chan *IntSet, dirtyOut chan *IntSet, visited []*IntSet) {
	go func() {
		expandedOut := make(chan *IntSet)
		newDirtyOut := make(chan *IntSet)
	
		for input := range in {
			log.Println("Starting slice run", len(input))
			for _, value := range input {
				<-simulRequestLimiter // Wait for next available 'request slot'
				log.Println("Took one, remaining:", len(simulRequestLimiter))
				go func(value interface{}) {
					log.Println("Sending value to mapper")
					expanded, dirty := mapper(value, visited)
					log.Println("Got value from mapper")
					expandedOut <- expanded
					log.Println("Sent on expanded")
					newDirtyOut <- dirty
					log.Println("Sent on dirty")
					simulRequestLimiter <- true // Mark one 'request slot' as available
					log.Println("Put one back on the list, remaining:", len(simulRequestLimiter))
				}(value)
			}

			midLevelSet := NewIntSet()
			dirtySet := NewIntSet()

			log.Println("Listening for results", len(input))
			for _ = range input {
				expanded := <-expandedOut
				midLevelSet.Union(expanded)
				log.Println("midLevelSet:", midLevelSet.Size())
				dirty := <-newDirtyOut
				dirtySet.Union(dirty)
			}

			out <- midLevelSet
			dirtyOut <- dirtySet
		}
	}()
}


func createSliceHandlers(num int, mapper func(interface{}, []*IntSet) (*IntSet, *IntSet), sliceInChan chan []interface{}, sliceOutChan chan *IntSet, sliceDirtyOutChan chan *IntSet, visited []*IntSet) {
	for i := 0; i < num; i++ {
		createSliceHandler(mapper, sliceInChan, sliceOutChan, sliceDirtyOutChan, visited)
	}
}


func createSearchHandler(mapper func(interface{}, []*IntSet) (*IntSet, *IntSet), prepper func(*IntSet) []interface{}, visited []*IntSet) (inChan, outChan, dirtyOutChan chan *IntSet) {
	inChan, outChan, dirtyOutChan = make(chan *IntSet), make(chan *IntSet), make(chan *IntSet)

	sliceInChan := make(chan []interface{}) // Every request funneled into one 'please' channel
	sliceOutChan := make(chan *IntSet) // Every response funneled into one 'finished' channel
	sliceDirtyOutChan := make(chan *IntSet) // Every response funneled into one 'finished' channel

	createSliceHandlers(NUM_INTERMEDIATES, mapper, sliceInChan, sliceOutChan, sliceDirtyOutChan, visited)

	go func() {
		for input := range inChan {
			log.Println("Starting search run")
			slices := partitionByNum(prepper(input), NUM_INTERMEDIATES)

			topLevelSet := NewIntSet()
			dirtySet := NewIntSet()

			for _, slice := range slices {
				go func(slice []interface{}) {
					sliceInChan <- slice
				}(slice)
			}

			for _ = range slices {
				results := <-sliceOutChan
				topLevelSet.Union(results)
				log.Println("topLevelSet:", topLevelSet.Size())
				dirty := <-sliceDirtyOutChan
				dirtySet.Union(dirty)
			}

			outChan <- topLevelSet
			dirtyOutChan <- dirtySet
		}
	}()

	return inChan, outChan, dirtyOutChan
}


func createSearchIterator() (inChan, outChan chan *IntSet, visited []*IntSet) {
	inChan, outChan = make(chan *IntSet), make(chan *IntSet)

	visited = make([]*IntSet, NUM_VISITED_SETS, NUM_VISITED_SETS)
	visited[MATCHES] = NewIntSet()
	visited[PLAYERS] = NewIntSet()

	go func() {
		leagueIn, leagueOut, leagueDirty := createSearchHandler(SearchPlayerLeague, InputPrepperLeague, visited)
		matchIn, matchOut, matchDirty := createSearchHandler(SearchPlayerMatch, InputPrepperMatch, visited)

		for input := range inChan {
			// Do league first, so league can weed out players of too-low tier
			leagueIn <- input
			outputLeague := <-leagueOut
			dirtyPlayers := <-leagueDirty

			fmt.Printf("\n Finished with leagues, got %d new players and %d low-tier players\n", outputLeague.Size(), dirtyPlayers.Size())
			input.IntersectInverse(dirtyPlayers) // Remove all dirty players (too low tier, etc.)

			matchIn <- input
			outputMatch := <-matchOut
			_ = <-matchDirty // Matches can't mark dirty players

			fmt.Printf("\n Finished with matches, got %d new players\n", outputMatch.Size())

			outputMatch.Union(outputLeague)

			outChan <- outputMatch
		}
	}()

	return inChan, outChan, visited
}


func search() {
	in, out, visited := createSearchIterator()

	initialSeeds := NewIntSet(51405, 10077)

	// fmt.Println("initialSeeds:", initialSeeds)

	newPlayers := initialSeeds

	var start time.Time

	for newPlayers.Size() > 0 {
		start = time.Now()

		visited[PLAYERS].Union(newPlayers)

		fmt.Printf("\nvisited (%d)\n", visited[PLAYERS].Size())
		fmt.Printf("newPlayers (%d)\n\n", newPlayers.Size())

		in <- newPlayers
		newPlayers = <-out

		fmt.Printf("\n\nIteration: %v\n", time.Since(start))
	}
}


func printTimeSince(startTime time.Time) {
    fmt.Println("Total elapsedTime:", time.Since(startTime))
}


func main() {
	// TImer for entire program
    defer printTimeSince(time.Now())


    // Seed random with nanoseconds
	rand.Seed(time.Now().UTC().UnixNano())


	// Set up cmd line flag
	var cpuprofile = flag.String("prof", "", "write cpu profile to file")

	// Handle cmd line flag behavior
    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }


    // Initialize simultaneous request limiter
    simulRequestLimiter = make(chan bool, MAX_SIMUL_REQUESTS)

    // Set up simultaneous request limiter with full allotment
    for i := 0; i < MAX_SIMUL_REQUESTS; i++ {
    	simulRequestLimiter <- true
    }


    // Continually print to stderr the number of goroutines (monitoring for leaks)
    go func() {
    	for _ = range time.Tick(5 * time.Second) {
    		log.Println("Number of goroutines:", runtime.NumGoroutine())
    	}
    }()


    // Set the number of parallel threads to use all CPUs
	fmt.Println("Default GOMAXPROCS:", runtime.GOMAXPROCS(runtime.NumCPU())) // Note: Setting, but returns the old for output
	fmt.Println("Running with GOMAXPROCS:", runtime.GOMAXPROCS(0))


	// Run
	search()
}