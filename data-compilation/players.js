var constants   = require('../helpers/constants.js'),
    fs          = require('fs');

function expandPlayersFromMatches(players, visitedPlayers, visitedMatches) {

}

function expandPlayersFromLeagues(players, visitedPlayers) {

}

function expandPlayers(players) {

}

function compilePlayersToFile(fileStream) {
    var players = constants.INITIAL_SEEDS;
}

function main() {
    var outFile = fs.createWriteStream('players.json');

    compilePlayersToFile(outFile);
}

main();