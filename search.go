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

func createSliceHandler(mapper func(interface{}, map[int]*IntSet) *IntSet, in chan []interface{}, out chan *IntSet, visited map[int]*IntSet) {
	go func() {
		for {
			input := <-in

			midLevelSet := NewIntSet()
			subOut := make(chan *IntSet)

			for _, value := range input {
				go func(value interface{}) {
					subOut <- mapper(value, visited)
				}(value)
			}

			for _ = range input {
				results := <-subOut
				midLevelSet.UnionWithout(results, visited[PLAYERS])
			}

			out <- midLevelSet
		}
	}()
}

func createSliceHandlers(num int, mapper func(interface{}, map[int]*IntSet) *IntSet, subInChan chan []interface{}, subOutChan chan *IntSet, visited map[int]*IntSet) {
	for i := 0; i < num; i++ {
		createSliceHandler(mapper, subInChan, subOutChan, visited)
	}
}

func createSearchHandler(mapper func(interface{}, map[int]*IntSet) *IntSet, prepper func(*IntSet) []interface{}, visited map[int]*IntSet) (inChan, outChan chan *IntSet) {
	inChan, outChan = make(chan *IntSet), make(chan *IntSet)

	subInChan := make(chan []interface{}) // Every request funneled into one 'please' channel
	subOutChan := make(chan *IntSet) // Every response funneled into one 'finished' channel

	createSliceHandlers(NUM_INTERMEDIATES, mapper, subInChan, subOutChan, visited)

	go func() {
		for {
			input := <-inChan
			slices := partitionByNum(prepper(input), NUM_INTERMEDIATES)
			// slices := partitionBySize(input, 3)
			superSet := NewIntSet()

			// fmt.Println("slices:", slices)

			for _, slice := range slices {
				go func(slice []interface{}) {
					subInChan <- slice
				}(slice)
			}

			for _ = range slices {
				results := <-subOutChan
				superSet.Union(results)
			}

			outChan <- superSet
		}
	}()

	return inChan, outChan
}

func createSearchIterator() (inChan, outChan chan *IntSet, visited map[int]*IntSet) {
	inChan, outChan = make(chan *IntSet), make(chan *IntSet)

	visited = make(map[int]*IntSet)
	visited[MATCHES] = NewIntSet()
	visited[PLAYERS] = NewIntSet()

	go func() {
		leagueIn, leagueOut := createSearchHandler(SearchPlayerLeague, InputPrepperLeague, visited)
		matchIn, matchOut := createSearchHandler(SearchPlayerMatch, InputPrepperMatch, visited)

		for {
			input := <-inChan

			// Do league first, so league can weed out players of too-low tier
			leagueIn <- input
			outputLeague := <-leagueOut

			matchIn <- input
			outputMatch := <-matchOut

			outputMatch.Union(outputLeague)

			outChan <- outputMatch
		}
	}()

	return inChan, outChan, visited
}

func search() {
	in, out, visited := createSearchIterator()

	initialSeeds := NewIntSet()
	initialSeeds.Add(51405)
	initialSeeds.Add(10077)

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

    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
	// defer un(trace())
	rand.Seed( time.Now().UTC().UnixNano())

	fmt.Println("Default GOMAXPROCS:", runtime.GOMAXPROCS(runtime.NumCPU())) // Note: Setting, but returns the old for output
	fmt.Println("Running with GOMAXPROCS:", runtime.GOMAXPROCS(0))

	search()
}