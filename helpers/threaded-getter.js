var promises = require('./promised.js');

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

process.on('message', function(obj) {
    console.log('Received:', obj);
    let iterable = obj.data;
    let resultHandler = eval(obj.resultHandler);
    let errorHandler = eval(obj.errorHandler);
    let limitSize = obj.limitSize;
    let promiseMapper = eval(obj.promiseMapper);

    return new Promise(function wrapper(resolve, reject) {
        let numTotal = iterable.length || iterable.size;
        let numActive = 1;      // Manually adjust for initial run
        let numReceived = -1;   // Manually adjust for initial run
        let iter = iterable[Symbol.iterator]();
        let elem = iter.next();

        let handleResponseAndSendNext = function() {
            --numActive;
            ++numReceived;

            process.stdout.write('\rReached ' + numReceived + ' / ' + numTotal + ' requests');

            if (numReceived >= numTotal) {
                process.stdout.write('\n');
                resolve();
            }
            else {
                while (numActive < limitSize && !elem.done) {
                    promiseMapper(elem.value)
                        .catch(errorHandler ? errorHandler : logErrorAndRethrow)
                        .then(resultHandler)
                        .then(handleResponseAndSendNext)
                        .catch(logErrorAndRethrow);
                    ++numActive;
                    elem = iter.next();
                }
            }
        };

        handleResponseAndSendNext();
    })
    .catch(logErrorAndRethrow);
});