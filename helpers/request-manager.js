'use strict';

var constants   = require('./constants.js'),
    fork        = require('child_process').fork;

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

function RequestManager(numThreads, maxRequests) {
    console.log('Concurrent requests:', maxRequests);
    console.log('Parallel threads:', numThreads);

    this.numThreads = numThreads;
    this.maxRequests = maxRequests;
}

RequestManager.prototype.get = function(iterable, mapFunc, resultHandler) {
    var self = this;

    return new Promise(function(resolve, reject) {
        let numRequests = (iterable.length || iterable.size);
        let numThreads = Math.min(self.numThreads, numRequests); // Make sure you don't have more threads than calls to make
        let threadSliceSize = Math.ceil((iterable.length || iterable.size) / numThreads);
        let results = [];
        let maxRequestsPerThread = Math.floor(self.maxRequests / numThreads);
        let numFinished = 0;
        let numReceived = 0; // Manually adjust for initial run

        // Edge case where the last thread wouldn't be handling any calls
        if ( numRequests === (threadSliceSize * (numThreads - 1)) ) {
            numThreads -= 1;
        }

        console.log('Handling', numRequests, 'over', numThreads, 'threads');

        let iter = iterable[Symbol.iterator]();
        let elem = iter.next();

        for (let i = 0; i < numThreads; ++i) {
            let newThread = fork(__dirname + '/getter-thread.js');

            let sliced = [];
            for (let i = 0; i < threadSliceSize && !elem.done; ++i) {
                sliced.push(elem.value);
                elem = iter.next();
            }

            newThread.send({ data: sliced.map(mapFunc), limitSize: maxRequestsPerThread, num: i });

            newThread.on('error', logErrorAndRethrow);

            newThread.on('message', function(msg) {
                if (msg.type === 'rec') {
                    ++numReceived;
                    resultHandler(msg.data);
                    process.stdout.write('\rReached: ' + numReceived + ' / ' + numRequests);
                }
                else if (msg.type === 'done') {
                    ++numFinished;

                    newThread.disconnect();

                    if (numFinished >= numThreads) {
                        process.stdout.write(' - Done.\n');
                        resolve();
                    }
                }
            });
        }
    });
    // body...
};

module.exports = RequestManager;