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

var cpuprofile = flag.String("prof", "", "write cpu profile to file")

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
	
		for {
			input := <-in

			for _, value := range input {
				go func(value interface{}) {
					expanded, dirty := mapper(value, visited)
					expandedOut <- expanded
					newDirtyOut <- dirty
				}(value)
			}

			midLevelSet := NewIntSet()
			dirtySet := NewIntSet()

			for _ = range input {
				expanded := <-expandedOut
				midLevelSet.Union(expanded)
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
		for {
			input := <-inChan
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

		for {
			input := <-inChan

			// Do league first, so league can weed out players of too-low tier
			leagueIn <- input
			outputLeague := <-leagueOut
			dirtyPlayers := <-leagueDirty

			input.IntersectInverse(dirtyPlayers) // Remove all dirty players (too low tier, etc.)

			matchIn <- input
			outputMatch := <-matchOut
			_ = <-matchDirty // Matches can't mark dirty players

			outputMatch.Union(outputLeague)

			outChan <- outputMatch
		}
	}()

	return inChan, outChan, visited
}

func search() {
	in, out, visited := createSearchIterator()

	initialSeeds := NewIntSet(51405, 10077)
	// initialSeeds.Add(51405)
	// initialSeeds.Add(10077)

	// fmt.Println("initialSeeds:", initialSeeds)

	newPlayers := initialSeeds

	var start time.Time

	for newPlayers.Size() > 0 {
		start = time.Now()

		visited[PLAYERS].Union(newPlayers) // TODO: Remove all low-skill players?

		fmt.Printf("\nvisited (%d)\n", visited[PLAYERS].Size())
		fmt.Printf("newPlayers (%d)\n\n", newPlayers.Size())

		in <- newPlayers
		newPlayers = <-out

		fmt.Printf("\n\nIteration: %v\n", time.Since(start))
	}
}

func trace() time.Time {
    return time.Now()
}
func un(startTime time.Time) {
    fmt.Println("Total elapsedTime:", time.Since(startTime))
}


func main() {
    defer un(trace())

	rand.Seed(time.Now().UTC().UnixNano())

    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

    go func() {
    	for _ = range time.Tick(5 * time.Second) {
    		log.Println("Number of goroutines:", runtime.NumGoroutine())
    	}
    }()

	fmt.Println("Default GOMAXPROCS:", runtime.GOMAXPROCS(runtime.NumCPU())) // Note: Setting, but returns the old for output
	fmt.Println("Running with GOMAXPROCS:", runtime.GOMAXPROCS(0))

	search()
}