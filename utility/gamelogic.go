package utility

import(
	// "fmt"
	"math/rand"
	"time"

	. "github.com/bawjensen/dataplay/constants"
)

func SearchPlayerMatch(player int) (expandedPlayers IntSet) {
	// fmt.Printf("Expanding %v via matches\n", player)

	expandedPlayers = NewIntSet(player)

	for i := 0; i < MAX_NUM_PER_MATCH; i++ {
		expandedPlayers.Add(RandomSummonerId())
	}

	// time.Sleep(250 * time.Millisecond)
	time.Sleep(1000 * time.Millisecond)
	
	// fmt.Printf("Expanded %d via matches: %v\n", player, expandedPlayers)

	return
}

func SearchPlayerLeague(player int) (expandedPlayers IntSet) {
	// fmt.Printf("Expanding %v via leagues\n", player)

	expandedPlayers = NewIntSet(player)

	for i := 0; i < MAX_NUM_PER_LEAGUE; i++ {
		expandedPlayers.Add(RandomSummonerId())
	}
	
	// time.Sleep(1000 * time.Millisecond)
	time.Sleep(2000 * time.Millisecond)

	// fmt.Printf("Expanded %d via leagues: %v\n", player, expandedPlayers)

	return
}

func RandomSummonerId() int {
	return rand.Intn(100)
}