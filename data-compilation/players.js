var promises    = require('../helpers/promised.js'),
    Queue       = require('../helpers/queue.js'),
    querystring = require('querystring');

// --------------------------------------- Global Variables -------------------------------------

var NOW             = (new Date).getTime();
var WEEK_AGO        = NOW - 604800000 / 7; // One week in milliseconds
var MATCHES_DESIRED = 100000;

// console.log('Time threshold of a week ago:', WEEK_AGO);

var API_KEY         = process.env.RIOT_KEY;
var RATE_LIMIT      = 100;
var INITIAL_SEEDS   = new Set([
    51405,          // C9 Sneaky
    // 492066,         // C9 Hai
    47585509        // CyclicSpec
]);

var matchListEndpoint   = 'https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/';
var matchEndpoint       = 'https://na.api.pvp.net/api/lol/na/v2.2/match/';
var leagueEndpoint      = 'https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/'

var matchListOptions = {
    'rankedQueues': 'RANKED_SOLO_5x5',
    'seasons': 'SEASON2015',
    'beginTime': WEEK_AGO,
    'api_key': API_KEY
};
var matchOptions = {
    'includeTimeline': 'false',
    'api_key': API_KEY
}
var leagueOptions = {
    'api_key': API_KEY
}

var matchListQuery  = '?' + querystring.stringify(matchListOptions);
var matchQuery      = '?' + querystring.stringify(matchOptions);
var leagueQuery     = '?' + querystring.stringify(leagueOptions);

// --------------------------------------- Helper Functions -------------------------------------

var tierChecker = new Set(['CHALLENGER', 'MASTER', 'DIAMOND', 'PLATINUM']);
function highEnoughTier(leagueDto) {
    return tierChecker.has(leagueDto.tier);
}

function logErrorAndRethrow(err) {
    console.log(err.stack);
    throw err;
}

// --------------------------------------- Main Functions ---------------------------------------

function getMatchesFromPlayers(visited, players) {
    console.log('Getting matches for', players.size, 'players');
    var matches = new Set();

    return promises.rateLimitedGet(players, RATE_LIMIT,
            function mapPlayer(summonerId) {
                return promises.persistentGet(matchListEndpoint + summonerId + matchListQuery);
            },
            function handleMatchList(matchList) {
                if (matchList.totalGames != 0) {
                    matchList.matches.forEach(function(matchListEntry) {
                        if (matchListEntry.platformId === 'NA1') {
                            matches.add(parseInt(matchListEntry.matchId));
                        }
                    });
                }
            }
        )
        .then(function() {
            return matches;
        });
}

function getPlayersFromMatches(visited, matches) {
    console.log('Getting players for', matches.size, 'matches');
    var newPlayers = new Set();

    return promises.rateLimitedGet(matches, RATE_LIMIT,
            function mapMatch(matchId) {
                return promises.persistentGet(matchEndpoint + matchId + matchQuery);
            },
            function handleMatch(match) {
                match.participantIdentities.forEach(function(pIdentity) {
                    var summonerId = parseInt(pIdentity.player.summonerId);
                    if ( !(visited.has(summonerId)) ) {
                        newPlayers.add(summonerId);
                        visited.add(summonerId);
                    }
                });
            }
        )
        .then(function() {
            return newPlayers;
        });
}

function expandPlayersFromMatches(visited, players) {
    return getMatchesFromPlayers(visited, players)
        .then(getPlayersFromMatches.bind(null, visited));
}

function expandPlayersFromLeagues(visited, players) {
    console.log('Getting leagues for', players.size, 'players');
    var newPlayers = new Set(); // All purely newly discovered players

    var groupedPlayers = [];

    let i = 0;
    var summonerGroup = [];
    for (let summonerId of players) {
        summonerGroup.push(summonerId);
        ++i;

        if (i >= 10) {
            groupedPlayers.push(summonerGroup);
            summonerGroup = [];
            i = 0;
        }
    }
    groupedPlayers.push(summonerGroup);

    return promises.rateLimitedGet(groupedPlayers, RATE_LIMIT,
            function mapPlayer(summonerIdList) {
                return promises.persistentGet(leagueEndpoint + summonerIdList.join() + leagueQuery);
            },
            function handleLeague(playerLeagueMap) {
                Object.keys(playerLeagueMap).forEach(function(summonerId) {
                    var leagueDtoList = playerLeagueMap[summonerId];

                    leagueDtoList.forEach(function(leagueDto) {
                        if (leagueDto.queue === 'RANKED_SOLO_5x5') {
                            if (highEnoughTier(leagueDto)) {
                                // console.log(leagueDto.entries.length, 'new players');
                                leagueDto.entries.forEach(function(leagueDtoEntry) {
                                    var summonerId = parseInt(leagueDtoEntry.playerOrTeamId);

                                    if ( !(visited.has(summonerId)) ) {
                                        newPlayers.add(summonerId);
                                        visited.add(summonerId);
                                    }
                                });
                            }
                            // else {
                            //     // console.log('Removing', summonerId);
                            //     newPlayers.delete(parseInt(summonerId));
                            // }
                        }
                    });
                });
            }
        )
        .then(function() {
            return newPlayers;
        });
}

function expandPlayers(visited, players) {
    return expandPlayersFromLeagues(visited, players)
        .then(expandPlayersFromMatches.bind(null, visited));
}

function compilePlayers() {
    var players = new Set(INITIAL_SEEDS);
    var visited = new Set();

    function loop() {
        console.log('visited: ', visited.size);
        console.log('players: ', players.size);

        return expandPlayers(visited, players)
            .then(function(newPlayers) {
                players = newPlayers;
                if (players.size) {
                    loop();
                }
            });
    }

    loop();
}

compilePlayers();

// function fetchEverything() {
//     return new Promise(function(resolve, reject) {
//         var oldMatches = new Set();

//         function loop(players) {
//             if (!players) return;

//             expandPlayersFromLeagues(players)
//                 .then(getMatchesFromPlayers)
//                 .then(function(matches) {
//                     // Removing old matches, adding new to old
//                     for (let match of matches) {
//                         if (oldMatches.has(match)) {
//                             matches.delete(match);
//                         }
//                         else {
//                             oldMatches.add(match);
//                         }
//                     }

//                     // Check if done
//                     if (oldMatches.size > MATCHES_DESIRED) {
//                         console.log('\rWe got to', MATCHES_DESIRED, 'matches');
//                         resolve();
//                         return; // Returning nothing breaks the chain
//                     }
//                     else {
//                         return matches;
//                     }
//                 })
//                 .then(getPlayersFromMatches)
//                 .then(loop)
//                 .catch(logErrorAndRethrow);
//         }

//         loop(INITIAL_SEEDS);
//     });
// }

// fetchEverything().then(function() { console.log('here'); }).catch(logErrorAndRethrow);
