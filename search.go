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

// ------------------------------------ Helper logic -----------------------------------------------

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

// ------------------------------------ Search logic -----------------------------------------------


func createSliceHandler(mapper func(interface{}, []*IntSet) *IntSet, in chan []interface{}, out chan *IntSet, visited []*IntSet) {
	go func() {
		expandedOut := make(chan *IntSet)
	
		for input := range in {
			for _, value := range input {
				go func(value interface{}) {
					expanded := mapper(value, visited)
					expandedOut <- expanded
				}(value)
			}

			midLevelSet := NewIntSet()

			for _ = range input {
				expanded := <-expandedOut
				midLevelSet.Union(expanded)
			}

			out <- midLevelSet
		}
	}()
}


func createSliceHandlers(num int, mapper func(interface{}, []*IntSet) *IntSet, visited []*IntSet) (sliceInChan chan []interface{}, sliceOutChan chan *IntSet) {
	sliceInChan = make(chan []interface{}) // Every request funneled into one 'please' channel
	sliceOutChan = make(chan *IntSet) // Every response funneled into one 'finished' channel

	for i := 0; i < num; i++ {
		createSliceHandler(mapper, sliceInChan, sliceOutChan, visited)
	}

	return sliceInChan, sliceOutChan
}


func createSearchHandler(mapper func(interface{}, []*IntSet) *IntSet, prepper func(*IntSet) []interface{}, visited []*IntSet) (inChan, outChan chan *IntSet) {
	inChan, outChan = make(chan *IntSet), make(chan *IntSet)

	sliceInChan, sliceOutChan := createSliceHandlers(NUM_INTERMEDIATES, mapper, visited)

	go func() {
		for input := range inChan {
			slices := partitionByNum(prepper(input), NUM_INTERMEDIATES)

			topLevelSet := NewIntSet()

			for _, slice := range slices {
				go func(slice []interface{}) {
					sliceInChan <- slice
				}(slice)
			}

			for _ = range slices {
				results := <-sliceOutChan
				topLevelSet.Union(results)
			}

			outChan <- topLevelSet
		}
	}()

	return inChan, outChan
}


func createSearchIterator() (inChan, outChan chan *IntSet, visited []*IntSet) {
	inChan, outChan = make(chan *IntSet), make(chan *IntSet)

	visited = make([]*IntSet, NUM_VISITED_SETS, NUM_VISITED_SETS)
	visited[MATCHES] = NewIntSet()
	visited[PLAYERS] = NewIntSet()

	go func() {
		// leagueIn, leagueOut := createSearchHandler(SearchPlayerLeague, InputPrepperLeague, visited)
		matchIn, matchOut := createSearchHandler(SearchPlayerMatch, InputPrepperMatch, visited)

		for input := range inChan {
			// Do league first, so league can weed out players of too-low tier
			// leagueIn <- input
			matchIn <- input

			// outputLeague := <-leagueOut
			outputMatch := <-matchOut

			// fmt.Printf("\n Leagues: got %d new players\n", outputLeague.Size())
			// fmt.Printf("\n Matches: got %d new players\n", outputMatch.Size())

			// outputMatch.Union(outputLeague)

			outChan <- outputMatch
		}
	}()

	return inChan, outChan, visited
}


func search() {
	in, out, visited := createSearchIterator()

	initialSeeds := NewIntSet(51405, 10077)

	newPlayers := initialSeeds

	var start time.Time

	for newPlayers.Size() > 0 {
		start = time.Now()

		fmt.Printf("\nvisited (%d)\n", visited[PLAYERS].Size())
		fmt.Printf("newPlayers (%d)\n\n", newPlayers.Size())

		in <- newPlayers

		visited[PLAYERS].Union(newPlayers)
		newPlayers = <-out

		log.Println("Number of goroutines after iteration:", runtime.NumGoroutine())

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


    // Continually print to stderr the number of goroutines (monitoring for leaks)
    go func() {
    	for _ = range time.Tick(2 * time.Second) {
    		log.Println("Number of goroutines:", runtime.NumGoroutine())
    	}
    }()


    // Set the number of parallel threads to use all CPUs
	fmt.Println("Default GOMAXPROCS:", runtime.GOMAXPROCS(runtime.NumCPU())) // Note: Setting, but returns the old for output
	fmt.Println("Running with GOMAXPROCS:", runtime.GOMAXPROCS(0))


	// Run
	search()
}