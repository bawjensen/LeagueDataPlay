'use strict';

var constants   = require('./constants.js'),
    fork        = require('child_process').fork;

function logErrorAndRethrow(err) {
    console.error(err.stack || err);
    throw err;
}

function RequestManager(numThreads, maxRequests) {
    console.log('Concurrent requests:', maxRequests);
    console.log('Parallel workers:', numThreads);

    this.numThreads = numThreads;
    this.maxRequests = maxRequests;

    this.workers = [];
    for (let i = 0; i < this.numThreads; ++i) {
        this.workers.push(fork(__dirname + '/getter-worker.js'));
    }
}

RequestManager.prototype.get = function(iterable, mapFunc, resultHandler) {
    var self = this;

    let numRequests = (iterable.length || iterable.size);
    let numThreads = Math.min(self.numThreads, numRequests); // Make sure you don't have more workers than calls to make
    // let workerSliceSize = Math.ceil((iterable.length || iterable.size) / numThreads);
    let results = [];
    let maxRequestsPerThread = Math.floor(self.maxRequests / numThreads);
    let numFinished = 0;
    let numReceived = 0; // Manually adjust for initial run
    let numOverloadedThreads = (numRequests % numThreads); // Number of workers that get n + 1 instead of n calls
    let workerSliceSize;

    // let numCalls = 0;

    // console.log('Handling', numRequests, 'over', numThreads, 'workers');

    let iter = iterable[Symbol.iterator]();
    let elem = iter.next();

    return Promise.all(
        this.workers.map(function(worker, workerIndex) {
            return new Promise(function(resolve, reject) {
                workerSliceSize = Math.floor(numRequests / numThreads) + (workerIndex < numOverloadedThreads ? 1 : 0);
                let sliced = [];
                for (let i = 0; i < workerSliceSize && !elem.done; ++i) {
                    sliced.push(elem.value);
                    elem = iter.next();
                }

                worker.send({ data: sliced.map(mapFunc), maxRequests: maxRequestsPerThread });

                // worker.on('error', logErrorAndRethrow); // Doesn't seem to do anything?

                worker.on('message', function(msg) {
                    if (msg.type === 'rec') {
                        ++numReceived;
                        resultHandler(msg.data);
                        process.stdout.write('\rReached: ' + numReceived + ' / ' + numRequests + ' (' + (numThreads - numFinished) + ' open)');
                    }
                    else if (msg.type === 'quit') {
                        console.log('Got a 403, stopping everything immediately');
                        process.exit();
                    }
                    else if (msg.type === 'done') {
                        ++numFinished;

                        // worker.disconnect();
                        // worker.kill(); // Is this necessary?

                        resolve();
                    }
                });
            });
        })
    )
    .then(function() {
        process.stdout.write(' - Done.\n');
    });
};

module.exports = RequestManager;