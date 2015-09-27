'use strict';

console.log('Loaded test1.js');

var constants = require('../helpers/constants.js'),
    RequestManager = require('./test2.js');

var reqMan = new RequestManager(2, { constants: constants });

reqMan.get(['haha', 'boohoo'], function(msg) {
    console.log(msg);
});