'use strict';

console.log('Loaded test3.js');

function deserializeFunction(funcStr) {
    console.log('funcStr', funcStr);
    return new Function('return ' + funcStr)();
}

function insertDependencies(dependencies) {
    for (var key in dependencies) {
        GLOBAL[key] = dependencies[key];
    }
}

function handleMapping(inputs, mapper) {
    return new Promise(function(resolve, reject) {
        for (let input of inputs) {
            mapper(input);
        }

        resolve();
    });
}

process.on('message', function(msg) {
    switch(msg.type) {
        case 'dep': // dependencies
            insertDependencies(msg.value);
            break;
        case 'map': // map input to output
            handleMapping(msg.inputs, deserializeFunction(msg.mapper)).then(function(ids) {
                process.send({ type: 'done', result: ids })
            });
            break;
        default:
            console.log('Message not understood:', msg);
            break;
    }
});

// console.log(process.env);

