package api

// ----------------------------------------- API response types ------------------------------------

type MatchlistResponse struct {
	Matches     []struct {
		// Champion int64
		// Lane     string
		MatchId     int64
		// PlatformId   string
		Queue       string
		Region      string
		// Role     string
		// Season       string
		// Timestamp    int64
	}
	// EndIndex int 
	// StartIndex   int 
	// TotalGames   int
}

type MatchResponse struct {
	// MapId                    int
	// MatchCreation            int64
	// MatchDuration            int64
	// MatchId                  int64
	// MatchMode                string
	// MatchType                string
	// MatchVersion         string
	ParticipantIdentities   []struct {
		// ParticipantId            int
		Player                  struct {
			// MatchHistoryUri          string
			// ProfileIcon              int
			SummonerId              int64
			// SummonerName         string
		}
	}
	// Participants         []struct {
	//  ChampionId                  int
	//  HighestAchievedSeasonTier   string
	//  Masteries                   []Mastery
	//  ParticipantId               int
	//  Runes                       []Rune
	//  Spell1Id                    int
	//  Spell2Id                    int
	//  Stats                       ParticipantStats
	//  TeamId                      int
	//  Timeline                    ParticipantTimeline
	// }
	// PlatformId               string
	QueueType               string
	// Region                   string
	// Season                   string
	// Teams                    []Team
	// Timeline             Timeline
}

type LeagueResponse map[string][]LeagueDto
type LeagueDto struct {
	Entries             []struct {
	//  Division            string
	//  IsFreshBlood        bool
	//  IsHotStreak         bool
	//  IsInactive          bool
	//  IsVeteran           bool
	//  LeaguePoints        int
	//  Losses              int
	//  MiniSeries          struct {
	//      Losses              int
	//      Progress            string
	//      Target              int
	//      Wins                int
	//  }
		PlayerOrTeamId      string
	//  PlayerOrTeamName    string
	//  Wins                int
	}
	// Name             string
	// ParticipantId        string
	Queue               string
	Tier                string
}

type LeagueEntryResponse map[string][]LeagueDto
// Uses same LeagueDto as LeagueResponse