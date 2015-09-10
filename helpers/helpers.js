function handleLeague(playerLeagueMap) {
    if (playerLeagueMap.err) {
        console.log('here');
        console.log(playerLeagueMap.data.stack);
    }

    if (!playerLeagueMap) return;

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
}