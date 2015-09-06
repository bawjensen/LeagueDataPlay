var fs          = require('fs'),
    os          = require('os'),
    promises    = require('../helpers/promised.js'),
    Queue       = require('../helpers/queue.js'),
    querystring = require('querystring');

// --------------------------------------- Global Variables -------------------------------------

var NOW             = (new Date).getTime();
var WEEK_AGO        = NOW - 604800000; // One week in milliseconds
var MATCHES_DESIRED = 100000;

// console.log('Time threshold of a week ago:', WEEK_AGO);

var API_KEY         = process.env.BAS_RIOT_KEY;
var DEFAULT_RATE_LIMIT = 500;
var NUM_THREADS     = os.cpus().length - 2;
var RATE_LIMIT      = process.argv[2] ? parseInt(process.argv[2]) : DEFAULT_RATE_LIMIT;
var INITIAL_SEEDS   = [
    51405           // C9 Sneaky
    // 492066,         // C9 Hai
    // 47585509        // CyclicSpec
];

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
function highEnoughTier(tier) {
    return tierChecker.has(tier);
}

function logErrorAndRethrow(err) {
    console.log(err.stack);
    throw err;
}
function logAndIgnore404(err) {
    if (err.http_code === 404) {
        console.log('\rIgnoring:', err.message);
    }
    else {
        logErrorAndRethrow(err);
    }
}

// --------------------------------------- Main Functions ---------------------------------------

function getMatchesFromPlayers(players) {
    console.log('Getting matches for', players.size, 'players');

    if (!players.size) return Promise.resolve();

    var matches = new Set();

    return promises.rateLimitedThreadedGet(players, NUM_THREADS, RATE_LIMIT,
        function mapPlayer(summonerId) {
            return { url: (matchListEndpoint + summonerId + matchListQuery), func: 'persistentGet' };
        },
        function handleMatchList(matchList) {
            if (!matchList) return;
            if (matchList.totalGames != 0) {
                matchList.matches.forEach(function(matchListEntry) {
                    if (matchListEntry.platformId === 'NA1') {
                        matches.add(parseInt(matchListEntry.matchId));
                    }
                });
            }
        },
        logAndIgnore404)
        .then(function() {
            return matches;
        });
    }

function getPlayersFromMatches(visited, newPlayers, matches) {
    console.log('Getting players for', matches.size, 'matches');

    if (!matches.size) return Promise.resolve();

    return promises.rateLimitedThreadedGet(matches, NUM_THREADS, RATE_LIMIT,
        function mapMatch(matchId) {
            return { url: (matchEndpoint + matchId + matchQuery), func: 'persistentGet' };
        },
        function handleMatch(match) {
            if (!match) return;
            match.participantIdentities.forEach(function(pIdentity) {
                var summonerId = parseInt(pIdentity.player.summonerId);
                if ( !(visited.has(summonerId)) ) {
                    newPlayers.add(summonerId); // Add so they're returned as result
                    visited.add(summonerId);
                }
            });
        },
        logAndIgnore404);
}

function expandPlayersFromMatches(visited, newPlayers, players) {
    return getMatchesFromPlayers(players)
        .then(getPlayersFromMatches.bind(null, visited, newPlayers));
}

function expandPlayersFromLeagues(visited, newPlayers, players) {
    console.log('Getting leagues for', players.size, 'players');

    if (!players.size) return Promise.resolve();

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
    if (i !== 0) {
        groupedPlayers.push(summonerGroup);
    }

    return promises.rateLimitedThreadedGet(groupedPlayers, NUM_THREADS, RATE_LIMIT,
        function mapPlayer(summonerIdList) {
            return { url: (leagueEndpoint + summonerIdList.join() + leagueQuery), func: 'persistentGet' };
        },
        function handleLeague(objectResult) {
            if (!objectResult) return;

            var playerLeagueMap = objectResult;

            // Object.keys(playerLeagueMap).forEach(function(summonerId) {
            Object.keys(playerLeagueMap).forEach(function(summonerId) {
                var leagueDtoList = playerLeagueMap[summonerId];

                if (!leagueDtoList) { // If the summoner wasn't in the returned league, they are unranked/unplaced
                    players.delete(parseInt(summonerId));
                    return;
                }

                leagueDtoList.forEach(function(leagueDto) {
                    if (leagueDto.queue === 'RANKED_SOLO_5x5') {
                        if (highEnoughTier(leagueDto.tier)) {
                            leagueDto.entries.forEach(function(leagueDtoEntry) {
                                let newSummonerId = parseInt(leagueDtoEntry.playerOrTeamId);

                                if ( !(visited.has(newSummonerId)) ) {
                                    newPlayers.add(newSummonerId);
                                    visited.add(newSummonerId);
                                }
                            });
                        }
                        else { // Summoner was too low tier to be considered
                            players.delete(parseInt(summonerId));
                        }
                    }
                });
            });
        },
        function catchBadRequests(err) {
            if (err.http_code === 404) {
                console.log('\rGot a full list of 404, removing all ids from players');
                let offenders = err.identifier;

                for (let summonerId of offenders) {
                    players.delete(summonerId)
                }
            }
            else {
                throw err;
            }
        }).catch(logErrorAndRethrow);
}

function expandPlayers(visited, newPlayers, players) {
    return expandPlayersFromLeagues(visited, newPlayers, players)
        .then(expandPlayersFromMatches.bind(null, visited, newPlayers, players));
}

function compilePlayers() {
    var outFile     = fs.createWriteStream('visited.json');
    var players     = new Set(INITIAL_SEEDS);
    var visited     = new Set();
    var newPlayers  = new Set();

    console.log('Rate limiting to:', RATE_LIMIT);
    console.log('Spreading over threads:', NUM_THREADS);

    outFile.write('[' + INITIAL_SEEDS.join());

    var promiseChain = Promise.resolve();

    function loop(initialRun) {
        if (!initialRun)
            players.forEach(function(summonerId) { outFile.write(',' + summonerId); });

        console.log('visited: ', visited.size);
        console.log('players: ', players.size);

        if (players.size) {
            promiseChain = promiseChain
                .then(expandPlayers.bind(null, visited, newPlayers, players))
                .then(function() {
                    players = new Set(newPlayers);
                    newPlayers.clear();
                })
                .then(loop);
        }
        else {
            outFile.write(']')
            promiseChain = promiseChain
                .then(function() {
                    var end = (new Date).getTime();
                    var minutes = (end - start) / 60000;
                    console.log('Took:', minutes, 'minutes');
                });
        }
    }

    loop(true);

    return promiseChain;
}

var start = NOW;
compilePlayers()
    .catch(logErrorAndRethrow);
