var promises = require('./promised.js');

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

process.on('message', function(obj) {
    // console.log('Received:', obj);
    let iterable = obj.data;
    let limitSize = obj.limitSize;
    let func = obj.func;

    return new Promise(function wrapper(resolve, reject) {
        let numTotal = iterable.length || iterable.size;
        let numActive = 1;      // Manually adjust for initial run
        let numReceived = -1;   // Manually adjust for initial run
        let iter = iterable[Symbol.iterator]();
        let elem = iter.next();
        let results = [];

        let handleResponseAndSendNext = function() {
            --numActive;
            ++numReceived;

            process.stdout.write('\rReached ' + numReceived + ' / ' + numTotal + ' requests');

            if (numReceived >= numTotal) {
                process.stdout.write('\n');
                resolve(results);
            }
            else {
                while (numActive < limitSize && !elem.done) {
                    promises[elem.value.func](elem.value.url)
                        .then(function(result) { results.push(result); })
                        .then(handleResponseAndSendNext)
                        .catch(logErrorAndRethrow);

                    ++numActive;
                    elem = iter.next();
                }
            }
        };

        handleResponseAndSendNext();
    })
    .catch(logErrorAndRethrow)
    .then(function(results) { process.send(results); });
});