'use strict';

function Node(value) {
    this.cargo = value;
    this.next = null;
}

function Queue(iterable) {
    this.head = null;
    this.tail = null;
    this.empty = true;

    if (iterable) {
        for (let item of iterable) {
            this.enqueue(item);
        }
        this.empty = false;
    }
}

Queue.prototype.enqueue = function(value) {
    var newNode = new Node(value);

    if (this.head === null) {
        this.head = newNode;
        this.tail = this.head;
        this.empty = false;
    }
    else {
        this.tail.next = newNode;
        this.tail = this.tail.next;
    }
}

Queue.prototype.dequeue = function() {
    if (this.head === null) throw new Error('Dequeueing an empty queue');

    var value = this.head.value;
    this.head = this.head.next;
    if (this.head === null) this.empty = true;
    return value;
}

module.exports = Queue;