'use strict';

var MongoClient = require('mongodb').MongoClient,
    request     = require('request');

function openDB(url) {
    return new Promise(function(resolve, reject) {
        MongoClient.connect(url, function(err, db) {
            if (err) { reject(Error(err)); }
            else { resolve(db); }
        });
    }).catch(logErrorAndRethrow);
}

function count(cursor) {
    return new Promise(function(resolve, reject) {
        cursor.count(function(err, count) {
            if (err) { reject(err); }
            else { resolve(count); }
        });
    });
}

function logErrorAndRethrow(err) {
    console.error(err.stack);
    throw err;
}

function CachingLayer(getFunc) {
    this.get = getFunc;
}

// CachingLayer.prototype.get = function(url) {
//     console.log('four');
//     return new Promise(function(resolve, reject) {
//         request.get(url, function(err, resp, body) {
//             if (err) { reject(Error(err)); }
//             else { resolve({ resp: resp, data: body, id: url}); }
//         });
//     });
// }

CachingLayer.prototype.dbFetch = function(url) {
    let self = this;
    return new Promise(function(resolve, reject) {
        self.db.collection('data').find({ _id: url }).toArray(function(err, result) {
            if (err) { reject(err); }
            else { resolve(result[0].body); }
        });
    });
}

CachingLayer.prototype.dbStoreAndReturn = function(obj) {
    let self = this;
    return new Promise(function(resolve, reject) {
        self.db.collection('data').insert({ _id: obj.id, body: obj.body }, function(err) {
            if (err) { reject(err); }
            else { resolve(obj.body); }
        });
    });
}

CachingLayer.prototype.fetch = function(url) {
    let self = this;

    let promiseChain = this.db ?
        Promise.resolve() :
        openDB('mongodb://localhost:27017/test').then(function(db) { self.db = db; });

    return promiseChain.then(function() {
        var cursor = self.db.collection('data').find({ _id: url });

        return count(cursor).then(function(count) {
            if (count === 0) {
                // console.log('Gotta fetch', url);
                return self.get(url).then(self.dbStoreAndReturn.bind(self));
            }
            else {
                // console.log('Already have', url);
                return self.dbFetch(url);
            }
        })
        .catch(logErrorAndRethrow);
    });
};

CachingLayer.prototype.end = function() {
    this.db.close();
}

module.exports = CachingLayer;