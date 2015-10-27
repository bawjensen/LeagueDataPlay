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

// ------------------------------------ Search logic -----------------------------------------------


func createSearchHandler(mapper func(interface{}, []*IntSet) (*IntSet, *IntSet), prepper func(*IntSet, []*IntSet) []interface{}, visited []*IntSet) (inChan, outChan, rejectOutChan chan *IntSet) {
	inChan, outChan, rejectOutChan = make(chan *IntSet), make(chan *IntSet), make(chan *IntSet)

	go func() {
		resultsChan := make(chan *IntSet)
		rejectsChan := make(chan *IntSet)

		for input := range inChan {
			prepped := prepper(input, visited)

			searchSet := NewIntSet()
			rejectSet := NewIntSet()

			for _, mapperInput := range prepped {
				go func(mapperInput interface{}) {
					results, rejects := mapper(mapperInput, visited)
					resultsChan <- results
					rejectsChan <- rejects
				}(mapperInput)
			}

			for _ = range prepped {
				searchSet.Union(<-resultsChan)
				rejectSet.Union(<-rejectsChan)
			}

			outChan <- searchSet
			rejectOutChan <- rejectSet
		}
	}()

	return inChan, outChan, rejectOutChan
}


func createSearchIterator() (inChan, outChan chan *IntSet, visited []*IntSet) {
	inChan, outChan = make(chan *IntSet), make(chan *IntSet)

	visited = make([]*IntSet, NUM_VISITED_SETS, NUM_VISITED_SETS)
	visited[MATCHES] = NewIntSet()
	visited[PLAYERS] = NewIntSet()
	visited[LEAGUE_BY_PLAYERS] = NewIntSet()

	go func() {
		leagueIn, leagueOut, leagueReject := createSearchHandler(SearchPlayerLeague, InputPrepperLeague, visited)
		matchIn, matchOut, matchReject := createSearchHandler(SearchPlayerMatch, InputPrepperMatch, visited)

		for input := range inChan {
			// Do league first, so league can weed out players of too-low tier?
			leagueIn <- input
			outputLeague := <-leagueOut
			rejectLeague := <-leagueReject

			fmt.Printf("\nGot all league responses, rejecting: %d\n", rejectLeague.Size())

			input.Subtract(rejectLeague)

			matchIn <- input
			outputMatch := <-matchOut
			_ = <-matchReject

			// fmt.Printf("\n Leagues: got %d new players\n", outputLeague.Size())
			// fmt.Printf("\n Matches: got %d new players\n", outputMatch.Size())

			outputMatch.Union(outputLeague)

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

		visited[PLAYERS].Union(newPlayers)

		fmt.Printf("\nvisited[PLAYERS]:           %d\n", 	visited[PLAYERS].Size())
		fmt.Printf(  "visited[MATCHES]:           %d\n", 	visited[MATCHES].Size())
		fmt.Printf(  "visited[LEAGUE_BY_PLAYERS]: %d\n", 	visited[LEAGUE_BY_PLAYERS].Size())
		fmt.Printf(  "newPlayers:                 %d\n\n", 	newPlayers.Size())

		in <- newPlayers
		newPlayers = <-out

		log.Println("Number of goroutines after iteration:", runtime.NumGoroutine())

		fmt.Printf("\nIteration: %v\n", time.Since(start))
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


    // Set the number of parallel threads to use all CPUs
	fmt.Println("Default GOMAXPROCS:", runtime.GOMAXPROCS(runtime.NumCPU())) // Note: Setting, but returns the old for output
	fmt.Println("Running with GOMAXPROCS:", runtime.GOMAXPROCS(0))


	// Run
	search()
}