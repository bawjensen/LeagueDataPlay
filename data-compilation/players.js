var promises    = require('../helpers/promised.js'),
    querystring = require('querystring');

// --------------------------------------- Global Variables -------------------------------------

var NOW = (new Date).getTime();
var WEEK_AGO = NOW - 604800000 / 7; // One week in milliseconds

// console.log('Time threshold of a week ago:', WEEK_AGO);

var API_KEY = process.env.RIOT_KEY;
var RATE_LIMIT = 100;
var INITIAL_SEEDS = [
    51405,          // C9 Sneaky
    // 492066,         // C9 Hai
    47585509        // CyclicSpec
];

var matchListEndpoint = 'https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/';
var matchEndpoint = 'https://na.api.pvp.net/api/lol/na/v2.2/match/';
var leagueEndpoint = 'https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/'

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

function getMatchesFromPlayers(players) {
    console.log('Getting matches for', players.length, 'players');
    // console.log(players);
    var matches = new Set();

    return promises.rateLimitedGet(players, RATE_LIMIT,
            function mapPlayer(summonerId) {
                return promises.persistentGet(matchListEndpoint + summonerId + matchListQuery);
            },
            function handleMatchList(matchList) {
                if (matchList.totalGames != 0) {
                    matchList.matches.forEach(function(matchListEntry) {
                        if (match.platformId !== 'NA1') {
                            console.error('Whoops!');
                        }
                        matches.add(parseInt(matchListEntry.matchId));
                    });
                }
            }
        )
        .then(function() {
            var arrayMatches = [];
            matches.forEach(function(matchId) { arrayMatches.push(matchId); });
            return arrayMatches;
        })
        .catch(logErrorAndRethrow);
}

function getPlayersFromMatches(matches) {
    console.log('Getting players for', matches.length, 'matches');
    var players = new Set();

    return promises.rateLimitedGet(matches, RATE_LIMIT,
            function mapMatch(matchId) {
                return promises.persistentGet(matchEndpoint + matchId + matchQuery);
            },
            function handleMatch(match) {
                match.participantIdentities.forEach(function(pIdentity) {
                    players.add(parseInt(pIdentity.player.summonerId));
                });
            }
        )
        .then(function() {
            var arrayPlayers = [];
            players.forEach(function(summonerId) {
                arrayPlayers.push(summonerId);
            });
            return arrayPlayers;
        });
}

function getLeaguesFromPlayersAndExpand(players) {
    console.log('Getting leagues for', players.length, 'players');
    var expandedPlayers = new Set(players); // start the larger set off with the existing people

    var groupedPlayers = [];

    for (var i = 0, l = players.length; i < l; i += 10) { // 10 is maximum # of summoners at once
        groupedPlayers.push(players.slice(i, i + 10));
    }

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
                                    expandedPlayers.add(parseInt(leagueDtoEntry.playerOrTeamId));
                                });
                            }
                            else {
                                // console.log('Removing', summonerId);
                                expandedPlayers.delete(parseInt(summonerId));
                            }
                        }
                    });
                });
            }
        )
        .then(function() {
            var arrayPlayers = [];
            expandedPlayers.forEach(function(summonerId) {
                // console.log(summonerId);
                arrayPlayers.push(summonerId);
            });
            return arrayPlayers;
        });
}

getLeaguesFromPlayersAndExpand(INITIAL_SEEDS)
    .then(getMatchesFromPlayers)
    .then(getPlayersFromMatches)
    .then(getLeaguesFromPlayersAndExpand)
    .then(getMatchesFromPlayers)
    // .then(getPlayersFromMatches)
    // .then(getLeaguesFromPlayersAndExpand)
    .then(function(results) {
        // results.forEach(function(each) { console.log(each); });
        // console.log(results);
        console.log(results.length);
    })
    .catch(function(err) {
        console.log(err.stack);
    });