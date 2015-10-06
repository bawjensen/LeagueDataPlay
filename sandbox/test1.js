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



function getCallback(url, resolve, reject, err, resp, body) {
    if (err) {
        reject(err);
    }
    else {
        switch(resp.statusCode) {
            case 200:
                // console.log(body);
                resolve(body);
                break;
            case 429:
                // console.error('Got rate limited');
                // setTimeout(function() {
                //     request.get(url, getCallback.bind(null, url, identifier, resolve, reject));
                // }, parseInt(resp.headers['retry-after']));
                var rateLimitError = new Error('Rate limit from Riot\'s API');
                rateLimitError.code = resp.statusCode;
                rateLimitError.time = parseInt(resp.headers['retry-after']);
                rateLimitError.url = url;
                reject(rateLimitError);
                break;
            case 500:
            case 503:
            case 504:
                // console.error('Got', resp.statusCode, 'code, retrying in 0.5 sec (', url, ')');
                setTimeout(function() {
                    request.get(url, getCallback.bind(null, url, resolve, reject));
                }, 500);
                break;
            case 404:
                let error = new Error('Resp code was 404: ' + url);
                error.http_code = 404;
                // error.identifier = identifier;
                reject(error);
                break;
            case 403:
                process.send({ type: 'quit' });
                process.exit();
                break;
            default:
                reject(Error('Unhandled resp statusCode: ' + resp.statusCode + '(' + url + ')'));
                break;
        }
    }
}
function get(url) {
    return new Promise(function(resolve, reject) {
            request.get(url, getCallback.bind(undefined, url, resolve, reject));
            // if (Math.random() < 0.1) { request.get(url, getCallback.bind(undefined, url, resolve, reject)); }
            // else { getCallback(url, resolve, reject, null, { statusCode: 429, headers: { 'retry-after': 500 } }, null, null); }
        })
        .catch(function catchEndOfInputError(err) {
            if (err instanceof SyntaxError) {
                console.log('\rIgnoring:', url, err);
            }
            else {
                throw err;
            }
        })
        .catch(function(err) {
            if (err.code === 'ECONNRESET' || err.code === 'ETIMEDOUT') {
                console.error('\rIssue with:', url, '\n', err);
                // return get(url);
            }
            throw err;
        });
}

function createMapper(idToUrlMapper, respToIdsMapper) {
    return get(idToUrlMapper).then(respToIdsMapper);
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

function idListToUrlMapper(summonerIdList){
    return constants.LEAGUE_ENDPOINT + summonerIdList.join() + constants.LEAGUE_QUERY;
    // console.log(constants.LEAGUE_ENDPOINT + summonerIdList.join() + constants.LEAGUE_QUERY);
}

function leagueToIdsMapper(playerLeagueMap) {
    if (!playerLeagueMap) { return; }

    var ids = [];

    // Object.keys(playerLeagueMap).forEach(function(summonerId) {
    Object.keys(playerLeagueMap).forEach(function(summonerId) {
        summonerId = parseInt(summonerId);

        ids.push(summonerId);
    });

    return ids;
}

function expandPlayersFromLeagues(currentPlayers, newPlayers, visitedPlayers) {
    console.log('Getting leagues for', currentPlayers.size, 'players');

    if (!currentPlayers.size) return Promise.resolve();

    return requestManager.get(partition(currentPlayers, 10), createMapper(idListToUrlMapper, leagueToIdsMapper));

    // return requestManager.get(partition(currentPlayers, 10),
    //     function mapPlayer(summonerIdList) {
    //         return constants.LEAGUE_ENDPOINT + summonerIdList.join() + constants.LEAGUE_QUERY;
    //         // console.log(constants.LEAGUE_ENDPOINT + summonerIdList.join() + constants.LEAGUE_QUERY);
    //     }, handleLeague);
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