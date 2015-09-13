'use strict';

var querystring = require('querystring');

var apiKey = process.env.RIOT_CHALLENGE_KEY;

// var weekAgo = (new Date).getTime() - 604800000; // One week in milliseconds
var weekAgo = 1441950745000 - 604800000; // One week in milliseconds

var matchListOptions = {
    'rankedQueues': 'RANKED_SOLO_5x5',
    'seasons': 'SEASON2015',
    'beginTime': weekAgo,
    'api_key': apiKey
};
var matchOptions = {
    'includeTimeline': 'false',
    'api_key': apiKey
};
var leagueOptions = {
    'api_key': apiKey
};

module.exports = {
    INITIAL_SEEDS: new Set([
        51405,      // C9 Sneaky
        65399098    // TIP Rush
    ]),
    MAX_REQUESTS: 1000,
    // MAX_REQUESTS: 10,
    API_KEY: apiKey,

    MATCHLIST_ENDPOINT:  'https://na.api.pvp.net/api/lol/na/v2.2/matchlist/by-summoner/',
    MATCH_ENDPOINT:      'https://na.api.pvp.net/api/lol/na/v2.2/match/',
    LEAGUE_ENDPOINT:     'https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/',


    MATCHLIST_QUERY: '?' + querystring.stringify(matchListOptions),
    MATCH_QUERY:     '?' + querystring.stringify(matchOptions),
    LEAGUE_QUERY:    '?' + querystring.stringify(leagueOptions)
};