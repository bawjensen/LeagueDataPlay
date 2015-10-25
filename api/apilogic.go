package api

import (
	// "crypto/tls"
	"encoding/json"
	// "errors"
	// "fmt"
	"io"
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
)

// ----------------------------------------- Globals -----------------------------------------------

var client *http.Client
const (
	LIMIT_5XX = 5 // How many 5XX's we will allow before reacting
	SLEEP_5XX = 5 * time.Second // How long to sleep after every 5XX, before retrying request
	TIMEOUT_5XX = 30 * time.Second // How to long to take a timeout of everything after getting too many 5XX's
)

// ----------------------------------------- Helper logic ------------------------------------------

var tierChecker map[string]bool = map[string]bool{
	"CHALLENGER": 	true,
	"MASTER": 		true,
	"DIAMOND": 		true,
	"PLATINUM": 	true,
	"GOLD": 		false,
	"SILVER": 		false,
	"BRONZE": 		false,
}
func highEnoughTier(tierStr string) bool {
	_, ok := tierChecker[tierStr]
	return ok
}

func init() {
	// Set up client for HTTP gets
	tr := &http.Transport{
		MaxIdleConnsPerHost: MAX_SIMUL_REQUESTS,
	}
	client = &http.Client{Transport: tr}
}

// ----------------------------------------- General logic -----------------------------------------

func getJson(urlString string, data interface{}) {
	var resp *http.Response
	var err error

	<-simulRequestLimiter // Wait for next available 'request slot'

	got404 := false
	num5XX := 0
	gotResp := false
	for !gotResp {
		ratethrottle.Wait()

		eventReportChan <- REQUEST_SEND_EVENT

		// req, _ := http.NewRequest("GET", urlString, nil)
		// req.Header.Add("Connection", "keep-alive")
		// resp, err = client.Do(req)
		
		resp, err = client.Get(urlString)

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
				eventReportChan <- RESET_EVENT
			} else {
				log.Println("err:", err)
				eventReportChan <- UNKNOWN_ERROR_EVENT
			}
		} else {
			switch resp.StatusCode {
			case 200:
				gotResp = true

			case 429:
				sleepTimeSlice := resp.Header["Retry-After"]
				if len(sleepTimeSlice) > 0 {
					eventReportChan <- USER_RATE_LIMIT_EVENT
					sleep, _ := strconv.Atoi(sleepTimeSlice[0])
					time.Sleep(time.Duration(sleep))
				} else {
					eventReportChan <- SERV_RATE_LIMIT_EVENT
					if len(resp.Header["X-Rate-Limit-Type"]) > 0 {
						log.Println("Service 429?:", resp.Header)
					}
				}
				got404 = false // If a 429 follows a 404, don't mark the 404 as 'two consequtive'

			case 404:
				log.Println(resp.StatusCode, "-", urlString)
				if got404 {
					// Note not Fatal
					log.Println("Two 404's on this one url: ", resp.StatusCode, " ", urlString)
					gotResp = true
				}
				got404 = true

			case 500, 503, 504:
				eventReportChan <- SERVER_ERROR_EVENT
				if num5XX > LIMIT_5XX {
					log.Println(LIMIT_5XX, " 5XX's on this one url (", resp.StatusCode, "): ", urlString)
					ratethrottle.Sleep(TIMEOUT_5XX)
				}
				num5XX++
				time.Sleep(SLEEP_5XX)

			default:
				log.Fatal("Got ", resp.StatusCode, " with: ", urlString)
			}

			if !gotResp {
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
			}
		}
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(data)

	if decoder.More() {
		io.Copy(ioutil.Discard, resp.Body)
	}
	resp.Body.Close()

	eventReportChan <- REQUEST_SUCCESS_EVENT
	simulRequestLimiter <- signal{} // Mark one 'request slot' as available
}

// ----------------------------------------- Match logic -------------------------------------------

func InputPrepperMatch(players *IntSet, visited []*IntSet) (sliced []interface{}) {
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

func SearchPlayerMatch(iPlayer interface{}, visited []*IntSet) (expandedPlayers *IntSet) {
	player := iPlayer.(int64)

	expandedPlayers = NewIntSet()

	var matchlistData MatchlistResponse
	getJson(createMatchlistUrl(player), &matchlistData)

	ch := make(chan *IntSet)
	activeMatches := 0
	for _, match := range matchlistData.Matches {
		if match.Region == "NA" && match.Queue == DESIRED_QUEUE {
			activeMatches++
			go func(matchId int64) {
				newIds := NewIntSet()

				if (!visited[MATCHES].Has(matchId)) {
					visited[MATCHES].Add(matchId)

					var matchData MatchResponse
					getJson(createMatchUrl(matchId), &matchData)

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
	
	return expandedPlayers
}

// ----------------------------------------- League logic ------------------------------------------

func InputPrepperLeague(players *IntSet, visited []*IntSet) (sliced []interface{}) {
	numMaxSlices := int(math.Ceil(float64(players.Size()) / float64(PLAYERS_PER_LEAGUE_CALL)))
	sliced = make([]interface{}, 0, numMaxSlices)

	j := 0
	var slice []int64
	for value := range players.Values() {
		if !visited[LEAGUE_BY_PLAYERS].Has(value) {
			if j >= PLAYERS_PER_LEAGUE_CALL { // If you've finished a slice, insert and continue
				// sliced[i] = slice
				sliced = append(sliced, slice)
				j = 0
			}

			if j == 0 { // Starting new slices
				slice = make([]int64, 0, PLAYERS_PER_LEAGUE_CALL)
			}

			slice = append(slice, value)
			j++
		}
	}

	// Leftover
	// sliced[i] = slice
	sliced = append(sliced, slice)

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

func SearchPlayerLeague(iPlayers interface{}, visited []*IntSet) (expandedPlayers *IntSet) {
	players := iPlayers.([]int64)

	var leagueData LeagueResponse
	getJson(createLeagueUrl(players), &leagueData)

	expandedPlayers = NewIntSet()

	for playerId := range leagueData {
		for _, leagueDto := range leagueData[playerId] {
			if leagueDto.Queue == DESIRED_QUEUE {
				if highEnoughTier(leagueDto.Tier) {
					for _, entry := range leagueDto.Entries {
						id, _ := strconv.ParseInt(entry.PlayerOrTeamId, 10, 64)
						visited[LEAGUE_BY_PLAYERS].Add(id)
						if !visited[PLAYERS].Has(id) {
							expandedPlayers.Add(id)
						}
					}
				} /*else {
					id, _ := strconv.ParseInt(playerId, 10, 64)
					removedPlayers.Add(id)
				}*/
			}
		}
	}

	return expandedPlayers
}

func RandomSummonerId() int {
	return rand.Intn(1e5)
}