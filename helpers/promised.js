var request = require('request');

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

function persistentCallback(url, resolve, reject, err, resp, body) {
    if (err) {
        reject(err);
    }
    else if (resp.statusCode === 429) {
        // console.error('Got rate limited');
        setTimeout(function() {
            request.get(url, persistentCallback.bind(null, url, resolve, reject));
        }, parseInt(resp.headers['retry-after']));
    }
    else if (resp.statusCode === 503 || resp.statusCode === 500 || resp.statusCode === 504) {
        // console.error('Got', resp.statusCode, 'code, retrying in 0.5 sec (', url, ')');
        setTimeout(function() {
            request.get(url, persistentCallback.bind(null, url, resolve, reject));
        }, 500);
    }
    else if (resp.statusCode === 404) {
        reject(Error('Resp code was 404: ' + url));
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
            request.get(url, persistentCallback.bind(null, url, resolve, reject));
        })
        .then(JSON.parse)
        .then(function returnWithIdentifier(data) {
            return data ? // Return data+identifier, data or null
                        (identifier ?
                            { data: data, id: identifier } :
                            data)
                        : null;
        })
        .catch(function(err) {
            if (err.code === 'ECONNRESET' || err.code === 'ETIMEDOUT') {
                console.error('Issue with:', url, '\n', err);
                return persistentGet(url, identifier);
            }
            else {
                throw err;
            }
        });
}

function rateLimitedGet(list, limitSize, promiseMapper, resultHandler) {
    return new Promise(function wrapper(resolve, reject) {
        var numTotal = list.length ? list.length : list.size ? list.size : 0;
        var reportIncrement = Math.max(Math.round(numTotal / 100), 1);
        var currentPosition = 0;
        var numActive = 1; // Manually adjust for initial run
        var numReceived = -1; // Manually adjust for initial run

        var handleResponseAndSendNext = function(initialRun) {
            --numActive;
            ++numReceived;

            if ( (numReceived % reportIncrement === 0) || (numReceived === numTotal) ) {
                process.stdout.write('\rReached ' + numReceived + ' / ' + numTotal + ' requests');
            }

            if (currentPosition >= numTotal) {
                if (numActive === 0) {
                    console.log('');
                    resolve();
                }
                return;
            }

            while (numActive < limitSize && currentPosition < numTotal) {
                promiseMapper(list[currentPosition]).then(resultHandler).then(handleResponseAndSendNext).catch(logErrorAndRethrow);
                ++numActive;
                ++currentPosition;
            }
        };

        handleResponseAndSendNext(true);
    })
    .catch(logErrorAndRethrow);
}

module.exports = {
    get: get,
    persistentGet: persistentGet,
    rateLimitedGet: rateLimitedGet
};