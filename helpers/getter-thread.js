'use strict';

var request = require('request'),
    CachingLayer = require('./caching-layer.js');

var threadNum;
var cachingLayer;
var sleepTime;

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

function sleep(time) {
    if (time < 0) return Promise.resolve();
    return new Promise(function(resolve, reject) {
        setTimeout(resolve, time);
    });
}

function finishUp() {
    try {
        cachingLayer.end();
        process.send({ type: 'done' });
        // process.disconnect(); // Attempting to close children
    }
    catch (err) {
        console.log('Child of a terminated parent ending here');
        process.exit(); // Exit quietly if parent has ended
    }

}

function getCallback(url, resolve, reject, err, resp, body) {
    if (err) {
        reject(err);
    }
    else {
        switch(resp.statusCode) {
            case 200:
                // console.log(body);
                resolve({ body: body, id: url });
                break;
            case 429:
                // console.error('Got rate limited');
                // setTimeout(function() {
                //     request.get(url, getCallback.bind(null, url, identifier, resolve, reject));
                // }, parseInt(resp.headers['retry-after']));
                var rateLimitError = new Error('Rate limit from Riot\'s API');
                rateLimitError.code = resp.statusCode;
                rateLimitError.time = parseInt(resp.headers['retry-after']);
                rateLimitError.url = url;
                reject(rateLimitError);
                break;
            case 500:
            case 503:
            case 504:
                // console.error('Got', resp.statusCode, 'code, retrying in 0.5 sec (', url, ')');
                setTimeout(function() {
                    request.get(url, getCallback.bind(null, url, resolve, reject));
                }, 500);
                break;
            case 404:
                let error = new Error('Resp code was 404: ' + url);
                error.http_code = 404;
                // error.identifier = identifier;
                reject(error);
                break;
            default:
                reject(Error('Unhandled resp statusCode: ' + resp.statusCode + '(' + url + ')'));
                break;
        }
    }
}
function get(url) {
    return new Promise(function(resolve, reject) {
            request.get(url, getCallback.bind(undefined, url, resolve, reject));
            // if (Math.random() < 0.1) { request.get(url, getCallback.bind(undefined, url, resolve, reject)); }
            // else { getCallback(url, resolve, reject, null, { statusCode: 429, headers: { 'retry-after': 500 } }, null, null); }
        })
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

function fetch(url) {
    var promise;

    // if (Math.random() < 0.25) {
    //     console.log('Actually sending request');
        promise = cachingLayer.fetch(url)
            .then(JSON.parse);
    // }
    // else {
    //     console.log('\rFake rate limit');
    //     promise = new Promise(function(resolve, reject) {
    //         getCallback(url, resolve, reject, null, { statusCode: 429, headers: { 'retry-after': 2000 } }, null, null);
    //     });
    // }

    return promise;
}

function fetchAndSend(url) {
    // process.send({ type: 'req' });
    return fetch(url)
        .then(function(result) {
            /*console.log(result); */
            try {
                // console.log('\rSending "received" message');
                process.send({ type: 'rec', data: result });
            }
            catch (err) {
                finishUp();
            }
        })
        .catch(function catchRateLimit(err) {
            if (err.code === 429) {
                // console.log('\rCaught rate limit');
                sleepTime = (new Date).getTime() + err.time;
                return sleep(err.time)
                    .then(fetchAndSend.bind(undefined, err.url));
            }
            else {
                console.log('Unknown error:', err.stack)
                return { err: 'Unknown error', data: err.stack };
            }
        });
}

process.on('message', function(obj) {
    // console.log('Received:', obj);
    let iterable = obj.data;
    let maxRequests = obj.maxRequests;

    cachingLayer = new CachingLayer(get);
    
    threadNum = obj.num;

    return new Promise(function wrapper(resolve, reject) {
        let numTotal = iterable.length || iterable.size;
        let numActive = 1;      // Manually adjust for initial run
        let numReceived = -1;   // Manually adjust for initial run
        let iter = iterable[Symbol.iterator]();
        let elem = iter.next();
        let results = [];
        sleepTime = false;

        let rateLimitSleep = function() {
            if (sleepTime) {
                return sleep(sleepTime - (new Date).getTime())
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
                while (numActive < maxRequests && !elem.done) {
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