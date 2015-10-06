package main

import (
	"bytes"
	"fmt"
	"math/rand"
)

type IntSet struct {
	set map[int]bool
}

func NewIntSet() IntSet {
	return IntSet{make(map[int]bool)}
}

func (set IntSet) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[ ")
	for key := range set.set {
		buffer.WriteString(fmt.Sprint(key, " "))
	}
	buffer.WriteString("]")
	return buffer.String()
}

func (set *IntSet) Add(i int) {
	set.set[i] = true
}

func (set *IntSet) Union(other *IntSet) {
	for key := range other.set {
		set.Add(key)
	}
}

func (set *IntSet) Size() int {
	return len(set.set)
}

func getViaMatches(ch chan IntSet) {
	maxNumPerMatch := 10

	for {
		input := <- ch

		output := NewIntSet()

		for _ = range input.set {
			for i := 0; i < maxNumPerMatch; i++ {
				output.Add(rand.Intn(100))
			}
		}

		ch <- output
	}
}

func getViaLeague(ch chan IntSet) {
	maxNumPerLeague := 100

	for {
		input := <- ch

		fmt.Printf("Input: %v\n", input)
		// maxExpected := maxNumPerLeague * len(input)
		// ouput := make([]int, 0, maxExpected)
		output := NewIntSet()

		for _ = range input.set {
			for i := 0; i < maxNumPerLeague; i++ {
				output.Add(rand.Intn(100))
			}
		}

		ch <- output
	}
}

func main() {
	leagueChan := make(chan IntSet)
	matchesChan := make(chan IntSet)
	go getViaLeague(leagueChan)
	go getViaMatches(matchesChan)

	initialSeeds := NewIntSet()
	initialSeeds.Add(0)

	leagueChan <- initialSeeds
	matchesChan <- initialSeeds

	result1 := <- leagueChan
	result2 := <- matchesChan

	result := NewIntSet()
	result.Union(&result1)
	result.Union(&result2)

	// fmt.Printf("Result: %d %d %v %v\n", result1.Size(), result2.Size(), result1, result2)
	fmt.Printf("Result: %d %v\n", result.Size(), result)
}