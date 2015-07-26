var promises    = require('../helpers/promised.js'),
    querystring = require('querystring');

// promises.get('http://google.com').then(function(haha) { console.log(haha); }).catch(function(err) { console.log(err); });

// console.log(process.env.RIOT_KEY);

var apiKey = process.env.RIOT_KEY;
var initialSeeds = [
    51405
];

var matchListEndpoint = 'https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/';

var options = {
    'rankedQueues': 'RANKED_SOLO_5x5',
    'seasons': 'SEASON2015',
    'api_key': apiKey
};

var queryOptions = '?' + querystring.stringify(options);

function getMatchesFrom(players) {
    var matches = new Set();

    return Promise.all(
        players.map(function(summonerId) {
            return promises.persistentGet(matchListEndpoint + summonerId + queryOptions)
                .then(function(matchList) {
                    matchList.matches.forEach(function(matchListEntry) {
                        matches.add(matchListEntry.matchId);
                    });
                });
        })
    ).then(function() {
        return matches;
    });
}

getMatchesFrom(initialSeeds)
    .then(function(results) {
        results.forEach(function(each) { console.log(each); });
    })
    .catch(function(err) {
        console.log(err);
    });