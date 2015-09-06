var request = require('request'),
    fork    = require('child_process').fork;

// --------------------------------------- Helpers ---------------------------------------

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

// --------------------------------------- Main Functions --------------------------------

function get(url) {
    return new Promise(function(resolve, reject) {
        request.get(url, function(err, resp, body) {
            if (err)
                reject(Error(err));
            else
                resolve(body);
        });
    });
}

function persistentCallback(url, identifier, resolve, reject, err, resp, body) {
    if (err) {
        reject(err);
    }
    else if (resp.statusCode === 429) {
        // console.error('Got rate limited');
        // setTimeout(function() {
        //     request.get(url, persistentCallback.bind(null, url, identifier, resolve, reject));
        // }, parseInt(resp.headers['retry-after']));
        var rateLimitError = new Error('Got rate limited');
        rateLimitError.code = resp.statusCode;
        rateLimitError.time = parseInt(resp.headers['retry-after']);
        throw rateLimitError;
    }
    else if (resp.statusCode === 503 || resp.statusCode === 500 || resp.statusCode === 504) {
        // console.error('Got', resp.statusCode, 'code, retrying in 0.5 sec (', url, ')');
        setTimeout(function() {
            request.get(url, persistentCallback.bind(null, url, identifier, resolve, reject));
        }, 500);
    }
    else if (resp.statusCode === 404) {
        let error = new Error('Resp code was 404: ' + url);
        error.http_code = 404;
        error.identifier = identifier;
        reject(error);
    }
    else if (resp.statusCode !== 200) {
        reject(Error('Resp status code not 200: ' + resp.statusCode + '(' + url + ')'));
    }
    else {
        resolve(body);
    }
}
function persistentGet(url, identifier) {
    return new Promise(function get(resolve, reject) {
            request.get(url, persistentCallback.bind(null, url, identifier, resolve, reject));
        })
        .then(JSON.parse)
        .catch(function catchEndOfInputError(err) {
            if (err instanceof SyntaxError)
                console.log('\rIgnoring:', url, err);
            else
                throw err;
        })
        .then(function returnWithIdentifier(data) {
            return data ? // Return data+identifier, data or null
                        (identifier ?
                            { data: data, id: identifier } :
                            data)
                        : null;
        })
        .catch(function(err) {
            if (err.code === 'ECONNRESET' || err.code === 'ETIMEDOUT') {
                console.error('\rIssue with:', url, '\n', err);
                return persistentGet(url, identifier);
            }
            else {
                throw err;
            }
        });
}

function rateLimitedGet(iterable, limitSize, promiseMapper, resultHandler, errorHandler) {
    return new Promise(function wrapper(resolve, reject) {
        var isSet = (iterable instanceof Set) ? true : false;
        var isArray = !isSet;

        var numTotal = isArray ? iterable.length : iterable.size;
        var reportIncrement = Math.max(Math.round(numTotal / 100), 1);
        var numActive = 1;      // Manually adjust for initial run
        var numReceived = -1;   // Manually adjust for initial run
        var iter = iterable[Symbol.iterator]();
        var elem = iter.next();

        var handleResponseAndSendNext = function() {
            --numActive;
            ++numReceived;

            if ( (numReceived % reportIncrement === 0) || (numReceived === numTotal) ) {
                process.stdout.write('\rReached ' + numReceived + ' / ' + numTotal + ' requests');
            }

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
}

function rateLimitedThreadedGet(iterable, numThreads, limitSize, mapFunc, resultHandler, errorHandler) {
    return new Promise(function(resolve, reject) {
        let isArray = !!iterable.length;
        let numTotal = (iterable.length || iterable.size);
        numThreads = numThreads > numTotal ? numTotal : numThreads;
        let threadSliceSize = Math.round((iterable.length || iterable.size) / numThreads);
        let results = [];
        let numPerThread = Math.floor(limitSize / numThreads);
        let numReceived = 0;

        for (let i = 0; i < numThreads; ++i) {
            let newThread = fork(__dirname + '/../helpers/threaded-getter.js');

            let sliced;
            if (isArray) {
                sliced = iterable.slice(i * threadSliceSize, (i + 1) * threadSliceSize);
            }
            else {
                sliced = [];
                let iter = iterable[Symbol.iterator]();
                for (let i = 0; i < threadSliceSize; ++i) {
                    sliced.push(iter.next().value);
                }
            }

            newThread.send({ data: sliced.map(mapFunc), limitSize: numPerThread });

            newThread.on('error', logErrorAndRethrow);

            newThread.on('message', function(arrayResults) {
                ++numReceived;
                process.stdout.write('\rAnother thread finished: ' + numReceived);

                arrayResults.map(resultHandler);
                newThread.disconnect();

                if (numReceived >= numThreads) {
                    process.stdout.write(' - Done.\n');
                    resolve();
                }
            });
        }
    });
}

module.exports = {
    get: get,
    persistentGet: persistentGet,
    rateLimitedGet: rateLimitedGet,
    rateLimitedThreadedGet: rateLimitedThreadedGet
};