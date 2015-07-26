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
        console.log('Issue with: ' + url);
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
                promisePersistentGet(url, identifier);
            else
                throw err;
        });
}

module.exports = {
    get: get,
    persistentGet: persistentGet
};