'use strict';

console.log('Loaded test2.js');

var fork = require('child_process').fork;

class RequestManager {
    constructor(cores, dependencies) {
        this.cores = cores || 4;
        this.workers = [];

        console.log('dependencies:', dependencies);

        for (let i = 0; i < this.cores; ++i) {
            this.workers.push(fork(__dirname + '/test3.js'));
            this.workers[i].send({ type: 'dep', value: dependencies });
        }
    }

    get(inputs, mapper) {
        // console.log('id:');
        for (let i = 0; i < this.cores; ++i) {
            this.workers[i].send({ type: 'map', mapper: mapper, inputs: inputs.slice(i, i+1) });
        }
    }
};

module.exports = RequestManager;