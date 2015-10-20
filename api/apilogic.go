package api

import(
	// "crypto/tls"
	"encoding/json"
	// "errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	// "reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bawjensen/dataplay/ratethrottle"

	. "github.com/bawjensen/dataplay/utility"
	// . "github.com/bawjensen/dataplay/constants"
)

// ------------------------------------ Globals ----------------------------------------

var client *http.Client
var eventReportChan chan byte

// ------------------------------------ API response types -----------------------------

type MatchlistResponse struct {
	Matches		[]struct {
		// Champion	int64
		// Lane		string
		MatchId		int64
		// PlatformId	string
		Queue		string
		Region		string
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
	QueueType				string
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

const (
	REQUEST_EVENT = iota
	TIMEOUT_EVENT
	RATE_LIMIT_EVENT
)

// ------------------------------------ Helper logic -----------------------------------

func init() {
	eventReportChan = make(chan byte)

	tr := &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
	client = &http.Client{Transport: tr}

	go func() {
		var events [3]int

		var eventType byte

		for {
			eventType = <-eventReportChan

			events[eventType]++

			fmt.Printf("\rCurrently at %d requests, %d timeouts, %d rate limits", events[REQUEST_EVENT], events[TIMEOUT_EVENT], events[RATE_LIMIT_EVENT])
		}
	}()
}

// ------------------------------------ General logic ----------------------------------

func getJson(urlString string, data interface{}) {
	var resp *http.Response
	var err error
	gotResp := false
	for !gotResp {
		ratethrottle.Wait()

		// fmt.Println("Request to:", urlString)

		req, _ := http.NewRequest("GET", urlString, nil)
		req.Header.Add("Connection", "keep-alive")
		resp, err = client.Do(req)
		
		// fmt.Println("Got response")

		// resp, err = client.Get(urlString)

		if err != nil {
			wasTimeout := false
			wasReset := false
			switch err := err.(type) {
			case *net.OpError:
				oErr, _ := err.Err.(*net.OpError)
				wasReset = (oErr.Error() == syscall.ECONNRESET.Error())
			case *url.Error:
				nErr, ok := err.Err.(net.Error)
				wasTimeout = (ok && nErr.Timeout())
			case net.Error:
				wasTimeout = err.Timeout()
			}
			if wasTimeout {
				// fmt.Println("Timeout err:", err)
				eventReportChan <- TIMEOUT_EVENT
			} else if wasReset {
				// eventReportChan <- RESET_EVENT
			} else {
				// fmt.Println("wasn't timeout, time to fatal log")
				// log.Fatal(err)
				fmt.Println("Faking a fatal log, let's see what this does", err)
			}
		} else {
			defer resp.Body.Close()

			switch resp.StatusCode {
			case 200:
				gotResp = true
			case 429:
				sleepTimeSlice := resp.Header["Retry-After"]
				if len(sleepTimeSlice) > 0 {
					eventReportChan <- RATE_LIMIT_EVENT
					sleep, _ := strconv.Atoi(sleepTimeSlice[0])
					sleep += 1
					// fmt.Println("\rGot a 429 user-based rate limit, sleeping for", sleep)
					time.Sleep(time.Duration(sleep))
				}
			case 404:
				fmt.Println(resp.StatusCode, "-", urlString)
				// err = errors.New(fmt.Sprintf("Issue with: %s", urlString))
				gotResp = true
			}
		}
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(data)

	_, _ = ioutil.ReadAll(resp.Body)

	eventReportChan <- REQUEST_EVENT
}

// ------------------------------------ Match logic ------------------------------------

func InputPrepperMatch(players *IntSet) (sliced []interface{}) {
	sliced = make([]interface{}, 0, players.Size())
	for value := range players.Values() {
		sliced = append(sliced, value)
	}
	return sliced
}

func createMatchUrl(match int64) string {
	parts := []string{ MATCH_PREFIX, strconv.FormatInt(match, 10), "?api_key=", API_KEY }
	return strings.Join(parts, "")
}

func createMatchlistUrl(player int64) string {
	parts := []string{ MATCHLIST_PREFIX, strconv.FormatInt(player, 10), "?beginTime=", MATCH_BEGIN_TIME, "&api_key=", API_KEY }
	return strings.Join(parts, "")
}

func SearchPlayerMatch(iPlayer interface{}, visited map[int]*IntSet) (expandedPlayers *IntSet) {
	player := iPlayer.(int64)

	expandedPlayers = NewIntSet()

	var matchlistData MatchlistResponse
	matchlistUrl := createMatchlistUrl(player)
	getJson(matchlistUrl, &matchlistData)

	ch := make(chan *IntSet)
	activeMatches := 0
	for _, match := range matchlistData.Matches {
		if match.Region == "NA" && match.Queue == DESIRED_QUEUE {
			activeMatches++
			go func(matchId int64) {
				newIds := NewIntSet()

				if (!visited[MATCHES].Has(matchId)) {
					var matchData MatchResponse
					matchUrl := createMatchUrl(matchId)
					getJson(matchUrl, &matchData)

					if matchData.QueueType == DESIRED_QUEUE {
						for _, participant := range matchData.ParticipantIdentities {
							if !visited[PLAYERS].Has(participant.Player.SummonerId) {
								newIds.Add(participant.Player.SummonerId)
							}
						}
					}
				}

				ch <- newIds
			}(match.MatchId)
		}
	}

	for i := 0; i < activeMatches; i++ {
		results := <-ch
		expandedPlayers.Union(results)
	}

	// fmt.Printf("Got %d from %d matches\n", expandedPlayers.Size(), len(matchlistData.Matches))
	
	return expandedPlayers
}

// ------------------------------------ League logic -----------------------------------

func InputPrepperLeague(players *IntSet) (sliced []interface{}) {
	numSlices := int(math.Ceil(float64(players.Size()) / float64(PLAYERS_PER_LEAGUE_CALL)))
	sliced = make([]interface{}, numSlices, numSlices)

	i := 0
	j := 0
	var slice []int64
	for value := range players.Values() {
		if j >= PLAYERS_PER_LEAGUE_CALL { // If you've finished a slice, insert and continue
			sliced[i] = slice
			j = 0
			i++
		}

		if j == 0 { // Starting new slices
			slice = make([]int64, 0, PLAYERS_PER_LEAGUE_CALL)
		}

		// slice[j] = value
		slice = append(slice, value)
		j++
	}

	// Leftover
	sliced[i] = slice

	return sliced
}

func createLeagueUrl(players []int64) string {
	stringPlayers := make([]string, len(players), len(players))

	for i, id := range players {
		stringPlayers[i] = strconv.FormatInt(id, 10)
	}

	parts := []string{ LEAGUE_PREFIX, strings.Join(stringPlayers, ","), "?api_key=", API_KEY }
	return strings.Join(parts, "")
}

func SearchPlayerLeague(iPlayers interface{}, visited map[int]*IntSet) (expandedPlayers *IntSet) {
	players := iPlayers.([]int64)

	var leagueData LeagueResponse

	leagueUrl := createLeagueUrl(players)
	getJson(leagueUrl, &leagueData)

	expandedPlayers = NewIntSet()
	for playerId := range leagueData {
		for _, leagueDto := range leagueData[playerId] {
			if leagueDto.Queue == DESIRED_QUEUE {
				for _, entry := range leagueDto.Entries {
					id, _ := strconv.ParseInt(entry.PlayerOrTeamId, 10, 64)
					if !visited[PLAYERS].Has(id) {
						expandedPlayers.Add(id)
					}
				}
			}
		}
	}

	// fmt.Printf("Got %d from league\n", expandedPlayers.Size())

	return expandedPlayers
}

func RandomSummonerId() int {
	return rand.Intn(1e5)
}