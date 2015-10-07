package main

import (
	"fmt"
	// "math/rand"
	// "sync"
	. "github.com/bawjensen/dataplay/utility"
	. "github.com/bawjensen/dataplay/constants"
)

func partition(set IntSet, num int) [][]int {
	setSize := set.Size()
	sliceSize := setSize / num
	numOverloadedSlices := setSize % num

	slices := make([][]int, num, num)

	iterator := set.Values()

	for i := 0; i < num; i++ {
		numInSlice := sliceSize
		if i < numOverloadedSlices {
			numInSlice += 1
		}

		slice := make([]int, numInSlice, numInSlice)

		for j := 0; j < numInSlice; j++ {
			slice[j] = <-iterator
		}

		slices[i] = slice
	}

	return slices
}

func createSliceHandler(mapper func(int) IntSet) (in chan []int, out chan IntSet)  {
	in = make(chan []int)
	out = make(chan IntSet)

	go func() {
		for {
			input := <-in

			superSet := NewIntSet()

			for _, value := range input {
				set := mapper(value)
				superSet.Union(&set)
			}

			out <- superSet
		}
	}()

	return in, out
}

func createSearchHandler(mapper func(int) IntSet) (inChan, outChan chan IntSet) {
	inChan, outChan = make(chan IntSet), make(chan IntSet)

	subInChans := make([]chan []int, NUM_INTERMEDIATES, NUM_INTERMEDIATES)
	subOutChans := make([]chan IntSet, NUM_INTERMEDIATES, NUM_INTERMEDIATES)

	for i := 0; i < NUM_INTERMEDIATES; i++ {
		subInChans[i], subOutChans[i] = createSliceHandler(mapper)
	}

	go func() {
		for {
			input := <-inChan
			slices := partition(input, NUM_INTERMEDIATES)
			output := NewIntSet()

			for i := 0; i < NUM_INTERMEDIATES; i++ {
				go func(i int) {
					subInChans[i] <- slices[i]
				}(i)
			}

			for i := 0; i < NUM_INTERMEDIATES; i++ {
				results := <-subOutChans[i]
				output.Union(&results)
			}

			outChan <- output
		}
	}()

	return inChan, outChan
}

func createSearchIterator() (inChan, outChan chan IntSet) {
	inChan, outChan = make(chan IntSet), make(chan IntSet)

	go func() {
		leagueIn, leagueOut := createSearchHandler(SearchPlayerLeague)
		matchIn, matchOut := createSearchHandler(SearchPlayerMatch)

		for {
			input := <-inChan

			leagueIn <- input
			matchIn <- input

			outputLeague := <-leagueOut
			outputMatch := <-matchOut

			outputLeague.Union(&outputMatch)

			outChan <- outputLeague
		}
	}()

	return
}

func main() {
	// in, out := createSearchHandler(SearchPlayerLeague)
	in, out := createSearchIterator()

	initialSeeds := NewIntSet()
	// initialSeeds.Add(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	initialSeeds.Add(0, 1)

	// fmt.Println("initialSeeds:", initialSeeds)
	in <- initialSeeds
	// <-out
	results := <-out

	fmt.Println("results:", results)
	
	in <- results
	results = <-out

	fmt.Println("results:", results)
}