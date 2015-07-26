var request = require('request');

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
        console.log('Issue with:', url);
        console.log(err);
        reject(err);
    }
    else if (resp.statusCode === 429) {
        setTimeout(function() {
            request.get(url, persistentCallback.bind(null, url, resolve, reject));
        }, parseInt(resp.headers['retry-after']));
    }
    else if (resp.statusCode === 503 || resp.statusCode === 500 || resp.statusCode === 504) {
        console.log('Got', resp.statusCode, 'code, retrying in 0.5 sec');
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
    console.log(url);
    return new Promise(function get(resolve, reject) {
            request.get(url, persistentCallback.bind(null, url, resolve, reject));
        })
        .then(JSON.parse)
        .then(function returnWithIdentifier(data) {
            return data ?
                        (identifier ?
                            { data: data, id: identifier } :
                            data)
                        : null;
        })
        .catch(function(err) {
            if (err.code === 'ECONNRESET')
                return persistentGet(url, identifier);
            else
                throw err;
        });
}

function rateLimitedGet(list, limitSize, promiseMapper, resultHandler) {
    return new Promise(function wrapper(resolve, reject) {
        var numTotal = list.length ? list.length : list.size ? list.size : 0;
        var numActive = 0;
        var currentPosition = 0;

        var handleResponseAndSendNext = function() {
            --numActive;

            if (currentPosition >= numTotal) {
                if (numActive === 0) {
                    resolve();
                }
                return;
            }

            while (numActive < limitSize && currentPosition < numTotal) {
                promiseMapper(list[currentPosition]).then(resultHandler).then(handleResponseAndSendNext);
                ++numActive;
                ++currentPosition;

                if (currentPosition % limitSize === 0) {
                    console.log('Reached', currentPosition, 'requests, continuing');
                }
            }
        }

        while (numActive < limitSize && currentPosition < numTotal) {
            promiseMapper(list[currentPosition]).then(resultHandler).then(handleResponseAndSendNext);

            ++numActive;
            ++currentPosition;
        }
    })
    .catch(function(err) {
        console.log(err);
    });
}

module.exports = {
    get: get,
    persistentGet: persistentGet,
    rateLimitedGet: rateLimitedGet
};