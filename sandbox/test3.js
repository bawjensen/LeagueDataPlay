'use strict';

console.log('Loaded test3.js');

function insertDependencies(dependencies) {
    for (var key in dependencies) {
        GLOBAL[key] = dependencies[key];
    }
}

function handleMapping(inputs, mapper) {
    for (let input of inputs) {
        console.log(constants.MATCHLIST_ENDPOINT + input + constants.MATCHLIST_QUERY);
    }
    // console.log(constants.)
}

process.on('message', function(msg) {
    switch(msg.type) {
        case 'dep': // dependencies
            insertDependencies(msg.value);
            break;
        case 'map': // map input to output
            handleMapping(msg.inputs, msg.mapper);
            break;
        default:
            console.log('Message not understood:', msg);
            break;
    }
});

// console.log(process.env);

