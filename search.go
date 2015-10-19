package main

import (
	"fmt"
	// "math"
	"math/rand"
	"runtime"
	// "sync"
	"time"
	. "github.com/bawjensen/dataplay/api"
	. "github.com/bawjensen/dataplay/utility"
	// . "github.com/bawjensen/dataplay/constants"
)

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
		// leagueIn, leagueOut := createSearchHandler(SearchPlayerLeague, InputPrepperLeague, visited)
		matchIn, matchOut := createSearchHandler(SearchPlayerMatch, InputPrepperMatch, visited)

		for {
			input := <-inChan

			// leagueIn <- input
			matchIn <- input

			for value := range input.Values() {
				visited[PLAYERS].Add(value)
			}

			// outputLeague := <-leagueOut
			outputMatch := <-matchOut

			// fmt.Println("outputLeague:", outputLeague)
			// fmt.Println("outputMatch:", outputMatch)

			// outputLeague.Union(outputMatch)

			// outChan <- outputLeague
			outChan <- outputMatch
		}
	}()

	return inChan, outChan, visited
}

func search() {
	in, out, visited := createSearchIterator()

	initialSeeds := NewIntSet()
	// initialSeeds.Add(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	// initialSeeds.Add(0, 1, 2, 3)
	initialSeeds.Add(51405)
	initialSeeds.Add(10077)

	// fmt.Println("initialSeeds:", initialSeeds)

	newPlayers := initialSeeds

	var start time.Time

	for newPlayers.Size() > 0 {
	// for newPlayers.Size() < 100 {
		// fmt.Printf("visited (%d)\n", visited.Size())
		// fmt.Printf("newPlayers (%d)\n", newPlayers.Size())

		start = time.Now()

		in <- newPlayers
		newPlayers = <-out

		fmt.Printf("visited (%d)\n", visited[PLAYERS].Size())
		fmt.Printf("newPlayers (%d)\n", newPlayers.Size())

		fmt.Printf("Iteration: %v\n\n", time.Since(start))
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
	rand.Seed( time.Now().UTC().UnixNano())

	fmt.Println("Default GOMAXPROCS:", runtime.GOMAXPROCS(0))
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Running with GOMAXPROCS:", runtime.GOMAXPROCS(0))

	search()
}