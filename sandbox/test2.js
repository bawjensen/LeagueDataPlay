'use strict';

console.log('Loaded test2.js');

var fork = require('child_process').fork;

class RequestManager {
    constructor(cores, requestCap, dependencies) {
        let self = this;

        self.cores = cores || 4;
        self.requestCap = requestCap;
        self.workers = [];

        for (let i = 0; i < self.cores; ++i) {
            self.workers.push(fork(__dirname + '/test3.js'));
            self.workers[i].send({ type: 'dep', value: dependencies });
        }
    }

    get(inputs, mapper) {
        let self = this;

        return new Promise(function(resolve, reject) {
            var numActive = 0;

            // Create funciton here for correct closure env
            var handleWorkerMessage = function(msg) {
                switch (msg.type) {
                    case 'done':
                        console.log(msg.result);
                        --numActive;
                        break;
                    default:
                        console.log('Message not understood:', msg);
                        break;
                }

                if (numActive <= 0) {
                    console.log('done');
                    resolve();
                }
            };

            // for (let i = 0; i < self.cores; ++i) {
            let i = 0;
            for (let worker of self.workers) {
                worker.send({ type: 'map', mapper: mapper.toString(), inputs: inputs.slice(i, i+1) });

                worker.on('message', handleWorkerMessage);

                ++numActive;
                i++;
            }
        });
    }
}

module.exports = RequestManager;
