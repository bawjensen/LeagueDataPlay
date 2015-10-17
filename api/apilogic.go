package api

import(
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	// "reflect"
	"strconv"
	"strings"
	// "time"

	"github.com/bawjensen/dataplay/ratethrottle"

	. "github.com/bawjensen/dataplay/utility"
	// . "github.com/bawjensen/dataplay/constants"
)

// ------------------------------------ Globals ----------------------------------------

// client := &http.Client{}
var client http.Client

// ------------------------------------ API response types -----------------------------

type MatchlistResponse struct {
	Matches		[]struct {
		// Champion	int64
		// Lane		string
		MatchId		int64
		// PlatformId	string
		// Queue		string
		// Region		string
		// Role		string
		// Season		string
		// Timestamp	int64
	}
	// EndIndex	int 
	// StartIndex	int 
	// TotalGames	int
}

type MatchResponse struct {
	// MapId					int
	// MatchCreation			int64
	// MatchDuration			int64
	// MatchId					int64
	// MatchMode				string
	// MatchType				string
	// MatchVersion			string
	ParticipantIdentities	[]struct {
		// ParticipantId			int
		Player					struct {
			// MatchHistoryUri			string
			// ProfileIcon				int
			SummonerId				int64
			// SummonerName			string
		}
	}
	// Participants			[]struct {
	// 	ChampionId					int
	// 	HighestAchievedSeasonTier	string
	// 	Masteries					[]Mastery
	// 	ParticipantId				int
	// 	Runes						[]Rune
	// 	Spell1Id					int
	// 	Spell2Id					int
	// 	Stats						ParticipantStats
	// 	TeamId						int
	// 	Timeline					ParticipantTimeline
	// }
	// PlatformId				string
	// QueueType				string
	// Region					string
	// Season					string
	// Teams					[]Team
	// Timeline				Timeline
}

type LeagueResponse map[string][]LeagueDto
type LeagueDto struct {
	Entries				[]struct {
	// 	Division			string
	// 	IsFreshBlood		bool
	// 	IsHotStreak			bool
	// 	IsInactive			bool
	// 	IsVeteran			bool
	// 	LeaguePoints		int
	// 	Losses				int
	// 	MiniSeries			struct {
	// 		Losses				int
	// 		Progress			string
	// 		Target				int
	// 		Wins				int
	// 	}
		PlayerOrTeamId		string
	// 	PlayerOrTeamName 	string
	// 	Wins				int
	}
	// Name				string
	// ParticipantId		string
	Queue				string
	Tier				string
}

// ------------------------------------ General logic ----------------------------------

func getJson(url string, data interface{}) {
	ratethrottle.Wait()
	// fmt.Println("Sending a request:", url)
	// resp, err := http.Get(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	// req.Header.Add("Connection", "close")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if (resp.StatusCode != 200) {
		log.Fatal(http.StatusText(resp.StatusCode))
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(data)
}

// ------------------------------------ Match logic ------------------------------------

func InputPrepperMatch(players IntSet) (sliced []interface{}) {
	sliced = make([]interface{}, 0, players.Size())
	for value := range players.Values() {
		sliced = append(sliced, value)
	}
	return
}

func createMatchUrl(match int64) string {
	return MATCH_PREFIX + strconv.FormatInt(match, 10) + "?api_key=" + API_KEY
}

func createMatchlistUrl(player int) string {
	return MATCHLIST_PREFIX + strconv.Itoa(player) + "?beginTime=" + MATCH_BEGIN_TIME + "&api_key=" + API_KEY
}

func SearchPlayerMatch(iPlayer interface{}) (expandedPlayers IntSet) {
	player := iPlayer.(int)

	expandedPlayers = NewIntSet()

	var matchlistData MatchlistResponse
	matchlistUrl := createMatchlistUrl(player)
	getJson(matchlistUrl, &matchlistData)

	var matchData MatchResponse
	ch := make(chan IntSet)
	for _, match := range matchlistData.Matches {
		matchUrl := createMatchUrl(match.MatchId)
		go func(url string) {
			getJson(url, &matchData)

			newIds := NewIntSet()

			for _, participant := range matchData.ParticipantIdentities {
				newIds.Add(int(participant.Player.SummonerId)) // TODO: Do I need to handle int64?
			}

			ch <- newIds
		}(matchUrl)
	}

	for _ = range matchlistData.Matches {
		results := <-ch
		fmt.Printf("Got %d from matches\n", results.Size())
		expandedPlayers.Union(&results)
	}
	
	return
}

// ------------------------------------ League logic -----------------------------------

func InputPrepperLeague(players IntSet) (sliced []interface{}) {
	numSlices := int(math.Ceil(float64(players.Size()) / float64(PLAYERS_PER_LEAGUE_CALL)))
	sliced = make([]interface{}, numSlices, numSlices)

	i := 0
	j := 0
	var slice []int
	for value := range players.Values() {
		if j >= PLAYERS_PER_LEAGUE_CALL { // If you've finished a slice, insert and continue
			sliced[i] = slice
			j = 0
			i++
		}

		if j == 0 { // Starting new slices
			slice = make([]int, 0, PLAYERS_PER_LEAGUE_CALL)
		}

		// slice[j] = value
		slice = append(slice, value)
		j++
	}

	// Leftover
	sliced[i] = slice

	return
}

func createLeagueUrl(players []int) string {
	stringPlayers := make([]string, len(players), len(players))

	for i, id := range players {
		stringPlayers[i] = strconv.Itoa(id)
	}

	return LEAGUE_PREFIX + strings.Join(stringPlayers, ",") + "?api_key=" + API_KEY
}

func SearchPlayerLeague(iPlayers interface{}) (expandedPlayers IntSet) {
	players := iPlayers.([]int)

	var leagueData LeagueResponse

	leagueUrl := createLeagueUrl(players)
	getJson(leagueUrl, &leagueData)

	expandedPlayers = NewIntSet()
	for playerId := range leagueData {
		for _, leagueDto := range leagueData[playerId] {
			if leagueDto.Queue == DESIRED_QUEUE {
				for _, entry := range leagueDto.Entries {
					id, _ := strconv.Atoi(entry.PlayerOrTeamId)
					expandedPlayers.Add(id)
				}
			}
		}
	}

	fmt.Printf("Got %d from league", expandedPlayers.Size())

	return
}

func RandomSummonerId() int {
	return rand.Intn(1e5)
}