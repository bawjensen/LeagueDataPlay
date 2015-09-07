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
    let threadNum = obj.num;

    return new Promise(function wrapper(resolve, reject) {
        let numTotal = iterable.length || iterable.size;
        let numActive = 1;      // Manually adjust for initial run
        let numReceived = -1;   // Manually adjust for initial run
        let iter = iterable[Symbol.iterator]();
        let elem = iter.next();
        let results = [];
        let sleepTime = false;

        let rateLimitSleep = function() {
            if (sleepTime) {
                console.log('Sleeping it!');
                return promises.sleep(sleepTime)
                    .then(function() { sleepTime = false });
            }
            else {
                return Promise.resolve();
            }
        }

        let handleResponseAndSendNext = function() {
            --numActive;
            ++numReceived;

            if (numReceived !== 0) {
                process.send({ type: 'rec' });
            }

            if (numReceived >= numTotal) {
                // resolve(results);
                resolve();
            }
            else {
                while (numActive < limitSize && !elem.done) {
                    promises.persistentGet(elem.value)
                        .catch(function catchRateLimit(err) {
                            if (err.code === 429) {
                                sleepTime = err.time;
                                return promises.sleep(err.time)
                                    .then(promises.persistentGet.bind(null, err.url))
                                    .catch(catchRateLimit);
                            }
                            else {
                                console.log('Unknown error:', err.stack)
                                return { err: 'Unknown error', data: err };
                            }
                        })
                        // .then(function(result) { results.push(result); })
                        .then(function(result) { process.send({ type: 'rec', data: result }); })
                        .then(rateLimitSleep)
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
    .then(function() { process.send({ type: 'done' }); });
});