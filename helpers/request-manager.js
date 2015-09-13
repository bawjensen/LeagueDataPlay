'use strict';

var constants   = require('./constants.js'),
    fork        = require('child_process').fork;

function logErrorAndRethrow(err) {
    console.error(err.stack || err);
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
        // let threadSliceSize = Math.ceil((iterable.length || iterable.size) / numThreads);
        let results = [];
        let maxRequestsPerThread = Math.floor(self.maxRequests / numThreads);
        let numFinished = 0;
        let numReceived = 0; // Manually adjust for initial run
        let numOverloadedThreads = (numRequests % numThreads); // Number of threads that get n + 1 instead of n calls
        let threadSliceSize;

        // let numCalls = 0;

        // console.log('Handling', numRequests, 'over', numThreads, 'threads');

        let iter = iterable[Symbol.iterator]();
        let elem = iter.next();

        for (let i = 0; i < numThreads; ++i) {
            let newThread = fork(__dirname + '/getter-thread.js');

            threadSliceSize = Math.floor(numRequests / numThreads) + (i < numOverloadedThreads ? 1 : 0);
            let sliced = [];
            for (let i = 0; i < threadSliceSize && !elem.done; ++i) {
                sliced.push(elem.value);
                elem = iter.next();
            }

            newThread.send({ data: sliced.map(mapFunc), maxRequests: maxRequestsPerThread, num: i });

            newThread.on('error', logErrorAndRethrow); // Doesn't seem to do anything?

            newThread.on('message', function(msg) {
                if (msg.type === 'rec') {
                    ++numReceived;
                    resultHandler(msg.data);
                    process.stdout.write('\rReached: ' + numReceived + ' / ' + numRequests + ' (' + (numThreads - numFinished) + ' open)');
                }
                // else if (msg.type === 'req') {
                //     ++numCalls;
                // }
                else if (msg.type === 'done') {
                    ++numFinished;

                    // newThread.disconnect();
                    newThread.kill(); // Is this necessary?

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