package utility

import(
	"os"
	"strconv"
	"time"
)

const (
	// Based on search system
	NUM_SEARCHES = 2 // Leagues and Matches

	// Configurable
	NUM_INTERMEDIATES = 5
	MAX_NUM_PER_LEAGUE = 100
	MAX_NUM_PER_MATCH = 10

	MATCHLIST_PREFIX = "https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/"
	MATCH_PREFIX = "https://na.api.pvp.net/api/lol/na/v2.2/match/"
	LEAGUE_PREFIX = "https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/"
	DESIRED_QUEUE = "RANKED_SOLO_5x5"

	// Determined by API
	REQUEST_CAP = 2500 // requests
	REQUEST_PERIOD = 10 // seconds
	PLAYERS_PER_LEAGUE_CALL = 10
	ONE_DAY = 		24 * 60 * 60
	ONE_WEEK =  7 * 24 * 60 * 60
)

var (
	// Pseudo-Constants
	API_KEY = os.Getenv("BAS_RIOT_KEY")
	// WEEK_AGO = "1443653132000" // Hard-coded for now, computed later
	MATCH_BEGIN_TIME = strconv.FormatInt((time.Now().Unix() - ONE_DAY) * 1000, 10)
)