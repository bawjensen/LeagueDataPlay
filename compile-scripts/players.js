var promises    = require('../helpers/promised.js'),
    querystring = require('querystring');

// promises.get('http://google.com').then(function(haha) { console.log(haha); }).catch(function(err) { console.log(err); });

// console.log(process.env.RIOT_KEY);

var apiKey = process.env.RIOT_KEY;
var initialSeeds = [
    51405,          // C9 Sneaky
    492066,         // C9 Hai
    47585509        // CyclicSpec
];

var matchListEndpoint = 'https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/';
var matchEndpoint = 'https://na.api.pvp.net/api/lol/na/v2.2/match/'

var matchListOptions = {
    'rankedQueues': 'RANKED_SOLO_5x5',
    'seasons': 'SEASON2015',
    'api_key': apiKey
};
var matchOptions = {
    'includeTimeline': 'false',
    'api_key': apiKey
}

var matchListQuery = '?' + querystring.stringify(matchListOptions);
var matchQuery = '?' + querystring.stringify(matchOptions);

function getMatches(players) {
    var matches = new Set();

    return promises.rateLimitedGet(players, 1,
            function mapPlayer(summonerId) {
                return promises.persistentGet(matchListEndpoint + summonerId + matchListQuery);
            },
            function handleMatchList(matchList) {
                matchList.matches.slice(0,10).forEach(function(matchListEntry) {
                    matches.add(matchListEntry.matchId);
                });
            }
        )
        .then(function() {
            var arrayMatches = [];
            matches.forEach(function(matchId) { arrayMatches.push(matchId); });
            return arrayMatches;
        });
}

function getPlayers(matches) {
    var players = new Set();

    return promises.rateLimitedGet(matches, 50,
            function mapMatch(matchId) {
                return promises.persistentGet(matchEndpoint + matchId + matchQuery);
            },
            function handleMatch(match) {
                match.participantIdentities.forEach(function(pIdentity) {
                    players.add(pIdentity.player.summonerId);
                });
            }
        )
        .then(function() {
            var arrayPlayers = [];
            players.forEach(function(summonerId) { arrayPlayers.push(summonerId); });
            return arrayPlayers;
        });
}

getMatches(initialSeeds)
    .then(getPlayers)
    .then(function(results) {
        // results.forEach(function(each) { console.log(each); });
        console.log(results.length);
    })
    .catch(function(err) {
        console.log(err.stack);
    });