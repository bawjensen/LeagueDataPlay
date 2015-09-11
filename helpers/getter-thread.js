'use strict';


var request = require('request'),
    cachingLayer = require('./caching-layer.js');

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

function finishUp() {
    try {
        process.send({ type: 'done' });
    }
    catch (err) {
        console.log('Child of a terminated parent ending here');
        CachingLayer.end();
        process.exit(); // Exit quietly if parent has ended
    }

}

function cachedGet(url) {
    // return new Promise(function get(resolve, reject) {
    //     request.get(url, getCallback.bind(null, url, resolve, reject));
    // });
    return cachingLayer.fetch(url);
}

function getCallback(url, resolve, reject, err, resp, body) {
    if (err) {
        reject(err);
    }
    else if (resp.statusCode === 429) {
        // console.error('Got rate limited');
        // setTimeout(function() {
        //     request.get(url, getCallback.bind(null, url, identifier, resolve, reject));
        // }, parseInt(resp.headers['retry-after']));
        var rateLimitError = new Error('Rate limit from Riot\'s API');
        rateLimitError.code = resp.statusCode;
        rateLimitError.time = parseInt(resp.headers['retry-after']);
        rateLimitError.url = url;
        reject(rateLimitError);
    }
    else if (resp.statusCode === 503 || resp.statusCode === 500 || resp.statusCode === 504) {
        // console.error('Got', resp.statusCode, 'code, retrying in 0.5 sec (', url, ')');
        setTimeout(function() {
            request.get(url, getCallback.bind(null, url, resolve, reject));
        }, 500);
    }
    else if (resp.statusCode === 404) {
        let error = new Error('Resp code was 404: ' + url);
        error.http_code = 404;
        // error.identifier = identifier;
        reject(error);
    }
    else if (resp.statusCode !== 200) {
        reject(Error('Resp status code not 200: ' + resp.statusCode + '(' + url + ')'));
    }
    else {
        resolve(body);
    }
}
function get(url) {
    return cachedGet(url)
        .then(JSON.parse)
        .catch(function catchEndOfInputError(err) {
            if (err instanceof SyntaxError) {
                console.log('\rIgnoring:', url, err);
            }
            else {
                throw err;
            }
        })
        .catch(function(err) {
            if (err.code === 'ECONNRESET' || err.code === 'ETIMEDOUT') {
                console.error('\rIssue with:', url, '\n', err);
                return get(url);
            }
            else {
                throw err;
            }
        });
}

function fetchAndSend(url) {
    return get(url)
        .catch(function catchRateLimit(err) {
            if (err.code === 429) {
                sleepTime = err.time;
                return promises.sleep(err.time)
                    .then(promises.get.bind(null, err.url))
                    .catch(catchRateLimit);
            }
            else {
                console.log('Unknown error:', err.stack)
                return { err: 'Unknown error', data: err };
            }
        })
        // .then(function(result) { results.push(result); })
        .then(function(result) {
            /*console.log(result); */
            try {
                process.send({ type: 'rec', data: result });
            }
            catch (err) {
                finishUp();
            }
        });
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

            // if (numReceived !== 0) {
            //     process.send({ type: 'rec' });
            // }

            if (numReceived >= numTotal) {
                // resolve(results);
                resolve();
            }
            else {
                while (numActive < limitSize && !elem.done) {
                    fetchAndSend(elem.value)
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
    .then(finishUp);
});