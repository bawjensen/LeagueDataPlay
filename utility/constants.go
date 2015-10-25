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
	NUM_INTERMEDIATES = 2 // Number of workers per searching section (e.g. league/match)
	RATE_THROTTLE_GRANULARITY = 10.0 // Divide both time and requests by this value when throttling
	RATE_THROTTLE_BUFFER = 500 * time.Millisecond // seconds
	MAX_SIMUL_REQUESTS = 500

	MATCHLIST_PREFIX = "https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/"
	MATCH_PREFIX = "https://na.api.pvp.net/api/lol/na/v2.2/match/"
	LEAGUE_PREFIX = "https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/"
	DESIRED_QUEUE = "RANKED_SOLO_5x5"

	// Determined by API
	REQUEST_CAP = 3000 // requests
	REQUEST_PERIOD = 10 * time.Second // seconds

	PLAYERS_PER_LEAGUE_CALL = 10
	ONE_HOUR = 			 60 * 60
	ONE_DAY = 		24 * 60 * 60
	ONE_WEEK =  7 * 24 * 60 * 60
)

// Used, for the moment, solely for indexing into the visited sets
const (
	PLAYERS = iota // Player ids that have been visited
	MATCHES // Match ids that have been visited
	LEAGUE_BY_PLAYERS // Players whose league has been been visited - since leagues are not queried by their id directly
	// LEAGUES

	NUM_VISITED_SETS
)

var (
	// Pseudo-Constants
	API_KEY = os.Getenv("BAS_RIOT_KEY")
	MATCH_BEGIN_TIME = strconv.FormatInt((time.Now().Unix() - ONE_WEEK) * 1000, 10)
)
