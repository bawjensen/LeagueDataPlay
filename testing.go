package main

import(
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	// "net/http"

	. "github.com/bawjensen/dataplay/api"
	// . "github.com/bawjensen/dataplay/constants"
)

func main() {
	// var matchlistData MatchlistResponse
	// matchlistUrl := "https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/10077?beginTime=1443653132000&api_key=c595ebc0-f6d2-4cdc-98d0-85bbbebff054"
	// matchlistResp, _ := http.Get(matchlistUrl)
    // defer matchlistResp.Body.Close()
	// matchlistDecoder := json.NewDecoder(matchlistResp.Body)
	// matchlistDecoder.Decode(&matchlistData)
	// fmt.Println("matchlistData:", matchlistData)


	// var matchData MatchResponse
	// matchUrl := "https://na.api.pvp.net/api/lol/na/v2.2/match/1972683788?api_key=c595ebc0-f6d2-4cdc-98d0-85bbbebff054"
	// matchResp, _ := http.Get(matchUrl)
    // defer matchResp.Body.Close()
	// matchDecoder := json.NewDecoder(matchResp.Body)
	// matchDecoder.Decode(&matchData)
	// fmt.Println("matchData:", matchData)


	fmt.Println("results:", SearchPlayerMatch(10077))


	// var leagueData LeagueResponse
	// leagueUrl := "https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/10077?api_key=7c257b79-fb36-47a2-af6e-69782293e10e"
	// leagueResp, _ := http.Get(leagueUrl)
    // defer leagueResp.Body.Close()
	// leagueDecoder := json.NewDecoder(leagueResp.Body)
	// leagueDecoder.Decode(&leagueData)
	// fmt.Println("leagueData:", leagueData)


	// fmt.Println("results:", SearchPlayerLeague([]int{10077}))
}

// package main

// import (
//     "encoding/json"
//     "fmt"
//     "net/http"
// )

// func main() {
//     var data struct {
//         Items []struct {
//             Name              string
//             Count             int
//             Is_required       bool
//             Is_moderator_only bool
//             Has_synonyms      bool
//         }
//     }

//     r, _ := http.Get("https://api.stackexchange.com/2.2/tags?page=1&pagesize=100&order=desc&sort=popular&site=stackoverflow")
//     defer r.Body.Close()

//     dec := json.NewDecoder(r.Body)
//     dec.Decode(&data)

//     for _, item := range data.Items {
//         fmt.Printf("%s = %d\n", item.Name, item.Count)
//     }

// }