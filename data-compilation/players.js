'use strict';

var constants       = require('../helpers/constants.js'),
    fs              = require('fs'),
    RequestManager  = require('../helpers/request-manager.js'),
    os              = require('os');

// --------------------------------------- Global Variables -------------------------------------

var requestManager;
var outFile;

// --------------------------------------- Helper Functions -------------------------------------

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

function finishUp() {
    var minutes = ((new Date).getTime() - start) / 60000;
    console.log('Took:', minutes, 'minutes');
    outFile.write(']');
    outFile.close();
}

function partition(array, groupSize) {
    var groupedArray = [];

    var tempGroup = [];
    for (let summonerId of array) {
        tempGroup.push(summonerId);

        if (tempGroup.length >= groupSize) {
            groupedArray.push(tempGroup);
            tempGroup = [];
        }
    }
    if (tempGroup.length) {
        groupedArray.push(tempGroup);
    }

    return groupedArray;
}

var tierChecker = new Set(['CHALLENGER', 'MASTER', 'DIAMOND', 'PLATINUM']);
function highEnoughTier(tier) {
    return tierChecker.has(tier);
}

// --------------------------------------- Primary Functions ------------------------------------

function getMatchesFromPlayers(currentPlayers, visitedMatches) {
    console.log('Getting matches for', currentPlayers.size, 'players');

    if (!currentPlayers.size) return Promise.resolve();

    var matches = new Set();

    return requestManager.get(currentPlayers,
        function mapPlayer(summonerId) {
            return constants.MATCHLIST_ENDPOINT + summonerId + constants.MATCHLIST_QUERY;
        },
        function handleMatchList(matchList) {
            if (!matchList) return;

            if (matchList.err) {
                console.log('eleven');
                console.log(matchList.data.stack);
            }

            if (matchList.totalGames !== 0) {
                matchList.matches.forEach(function(matchListEntry) {
                    if (matchListEntry.platformId === 'NA1') {
                        let matchId = parseInt(matchListEntry.matchId);
                        if (!visitedMatches.has(matchId)) {
                            matches.add(matchId);
                        }
                    }
                });
            }
        })
        .then(function() {
            return matches;
        });
}

function getPlayersFromMatches(visitedPlayers, newPlayers, matches) {
    console.log('Getting players for', matches.size, 'matches');

    if (!matches.size) return Promise.resolve();

    return requestManager.get(matches,
        function mapMatch(matchId) {
            return constants.MATCH_ENDPOINT + matchId + constants.MATCH_QUERY;
        },
        function handleMatch(match) {
            if (!match) return;

            if (match.err) {
                console.log('twelve');
                console.log(match.data.stack);
            }

            match.participantIdentities.forEach(function(pIdentity) {
                var summonerId = parseInt(pIdentity.player.summonerId);
                if ( !(visitedPlayers.has(summonerId)) ) {
                    newPlayers.add(summonerId); // Add so they're returned as result
                    visitedPlayers.add(summonerId);
                }
            });
        });
}

function expandPlayersFromMatches(currentPlayers, newPlayers, visitedPlayers, visitedMatches) {
    return getMatchesFromPlayers(currentPlayers, visitedMatches)
        .then(getPlayersFromMatches.bind(null, visitedPlayers, newPlayers));
}

function expandPlayersFromLeagues(currentPlayers, newPlayers, visitedPlayers) {
    console.log('Getting leagues for', currentPlayers.size, 'players');

    if (!currentPlayers.size) return Promise.resolve();

    return requestManager.get(partition(currentPlayers, 10),
        function mapPlayer(summonerIdList) {
            return constants.LEAGUE_ENDPOINT + summonerIdList.join() + constants.LEAGUE_QUERY;
        },
        function handleLeague(playerLeagueMap) {
            if (!playerLeagueMap) { return; }

            if (playerLeagueMap.err) {
                console.log('thirteen');
                console.log(playerLeagueMap.data);
                return;
            }

            // Object.keys(playerLeagueMap).forEach(function(summonerId) {
            Object.keys(playerLeagueMap).forEach(function(summonerId) {
                var leagueDtoList = playerLeagueMap[summonerId];

                if (!leagueDtoList) { // If the summoner wasn't in the returned league, they are unranked/unplaced
                    currentPlayers.delete(parseInt(summonerId));
                    return;
                }

                leagueDtoList.forEach(function(leagueDto) {
                    if (leagueDto.queue === 'RANKED_SOLO_5x5') {
                        if (highEnoughTier(leagueDto.tier)) {
                            leagueDto.entries.forEach(function(leagueDtoEntry) {
                                let newSummonerId = parseInt(leagueDtoEntry.playerOrTeamId);

                                if ( !(visitedPlayers.has(newSummonerId)) ) {
                                    newPlayers.add(newSummonerId);
                                    visitedPlayers.add(newSummonerId);
                                }
                            });
                        }
                        else { // Summoner was too low tier to be considered
                            currentPlayers.delete(parseInt(summonerId));
                        }
                    }
                });
            });
        });
}

function expandPlayers(currentPlayers, newPlayers, visitedPlayers, visitedMatches) {
    return expandPlayersFromLeagues(currentPlayers, newPlayers, visitedPlayers)
        .then(expandPlayersFromMatches.bind(undefined, currentPlayers, newPlayers, visitedPlayers, visitedMatches));
}

function compilePlayersToFile() {
    var newPlayers = constants.INITIAL_SEEDS;
    var currentPlayers = new Set();
    var visitedMatches = new Set();
    var visitedPlayers = new Set();

    var promiseChain = Promise.resolve(); // Start off the chain

    outFile.write('[' + Array.from(constants.INITIAL_SEEDS).join());

    function iterateSearch() {
        console.log('visitedPlayers:', visitedPlayers.size, '- newPlayers:', newPlayers.size);

        if (newPlayers.size) { // Done when no new currentPlayers
            currentPlayers = new Set(newPlayers);
            newPlayers.clear();

            promiseChain = promiseChain
                .then(expandPlayers.bind(undefined, currentPlayers, newPlayers, visitedPlayers, visitedMatches))
                .then(function() {
                    newPlayers.forEach(function(summonerId) { outFile.write(',' + summonerId); });
                })
                .then(iterateSearch)
                .catch(logErrorAndRethrow);
        }
        else {
            promiseChain.then(finishUp).catch(logErrorAndRethrow);
        }
    }

    iterateSearch();

    return promiseChain;
}

function main() {
    outFile = fs.createWriteStream('players.json');

    var numThreads = Math.max(Math.floor(os.cpus().length * 0.75), 1);

    requestManager = new RequestManager(numThreads, constants.MAX_REQUESTS);

    return compilePlayersToFile();
}

var start = (new Date).getTime();
main()
    .catch(function(err) {
        console.log(err.stack || err);
    });