'use strict';

console.log('Loaded test1.js');

// --------------------------------------- Module Imports ---------------------------------------

var constants       = require('../helpers/constants.js'),
    fs              = require('fs'),
    RequestManager  = require('./test2.js'),
    os              = require('os');

// --------------------------------------- Global Variables -------------------------------------

var requestManager;
var outFile;
var start;

// --------------------------------------- Helper Functions -------------------------------------

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

function finishUp() {
    var minutes = (new Date().getTime() - start) / 60000;
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

function createMapper(idToUrlMapper, respToSetMapper) {
    // return request.get(idToUrlMapper)
}

// --------------------------------------- Primary Behavior Functions ---------------------------

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

function handleLeague(playerLeagueMap) {
    if (!playerLeagueMap) { return; }

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
}

function expandPlayersFromLeagues(currentPlayers, newPlayers, visitedPlayers) {
    console.log('Getting leagues for', currentPlayers.size, 'players');

    if (!currentPlayers.size) return Promise.resolve();

    return requestManager.get(partition(currentPlayers, 10),
        function mapPlayer(summonerIdList) {
            return constants.LEAGUE_ENDPOINT + summonerIdList.join() + constants.LEAGUE_QUERY;
            // console.log(constants.LEAGUE_ENDPOINT + summonerIdList.join() + constants.LEAGUE_QUERY);
        }, handleLeague);
}

function expandPlayers(currentPlayers, newPlayers, visitedPlayers, visitedMatches) {
    return expandPlayersFromLeagues(currentPlayers, newPlayers, visitedPlayers)
        .then(expandPlayersFromMatches.bind(undefined, currentPlayers, newPlayers, visitedPlayers, visitedMatches));
}

function compilePlayersToFile() {
    return new Promise(function(resolve, reject) {
        var newPlayers = constants.INITIAL_SEEDS;
        var currentPlayers;
        var visitedMatches = new Set();
        var visitedPlayers = new Set();

        outFile.write('[' + Array.from(constants.INITIAL_SEEDS).join());

        function iterateSearch() {
            console.log('visitedPlayers:', visitedPlayers.size, '- newPlayers:', newPlayers.size);

            if (newPlayers.size) { // Done when no new players
                currentPlayers = newPlayers;
                newPlayers = new Set();

                expandPlayers(currentPlayers, newPlayers, visitedPlayers, visitedMatches)
                    .then(function() {
                        newPlayers.forEach(function(summonerId) { outFile.write(',' + summonerId); });
                    })
                    .then(iterateSearch)
                    .catch(logErrorAndRethrow);
            }
            else {
                resolve();
            }
        }

        iterateSearch();
    });
}

function main() {
    outFile = fs.createWriteStream('./json/test.json');

    var numCpus = Math.max(Math.floor(os.cpus().length * 0.75), 1);

    requestManager = new RequestManager(numCpus, constants.MAX_REQUESTS, { constants: constants });

    return compilePlayersToFile();
}

start = new Date().getTime();
main()
    .catch(function(err) {
        console.log(err.stack || err);
    });