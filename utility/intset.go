package utility

import (
	"bytes"
	"fmt"
	"sync"

	// . "github.com/bawjensen/dataplay/constants"
)

type IntSet struct {
	sync.RWMutex // Allows infinte readers OR one writer to lock the intset for themselves
	set map[int64]bool
}

func NewIntSet(initElems ...int64) (set *IntSet) {
	set = &IntSet{ set: make(map[int64]bool) }
	for _, elem := range initElems {
		// set.set[elem] = true
		set.Add(elem)
	}
	return set
}

// func NewIntSetFromSlice(initElems []interface{}) (set *IntSet) {
// 	set = &IntSet{ set: make(map[int64]bool) }
// 	for _, elem := range initElems {
// 		// set.set[elem.(int64)] = true
// 		set.Add(elem)
// 	}
// 	return set
// }

func (self *IntSet) String() string {
	// No need to lock, because Values() does
	var buffer bytes.Buffer

	buffer.WriteString("IntSet [ ")
	for elem := range self.Values() {
		buffer.WriteString(fmt.Sprint(elem, " "))
	}
	buffer.WriteString("]")

	return buffer.String()
}

func (self *IntSet) Add(elems ...int64) {
	self.Lock()
	defer self.Unlock()

	for _, elem := range elems {
		self.set[elem] = true
	}
}

func (self *IntSet) Remove(elems ...int64) {
	self.Lock()
	defer self.Unlock()

	for _, elem := range elems {
		delete(self.set, elem)
	}
}

func (self *IntSet) Has(elem int64) bool {
	self.RLock()
	defer self.RUnlock()
	return self.set[elem]
}

func (self *IntSet) Union(other *IntSet) {
	for elem := range other.Values() {
		self.Add(elem)
	}
}

func (self *IntSet) UnionWithout(other *IntSet, exclude *IntSet) {
	// No need to lock, because Values(), Has and Add do
	for elem := range other.Values() {
		if !exclude.Has(elem) {
			self.Add(elem)
		}
	}
}

func (self *IntSet) IntersectInverse(other *IntSet) {
	for elem := range other.Values() {
		self.Remove(elem)
	}
}

func (self *IntSet) Size() int {
	self.RLock()
	defer self.RUnlock()

	return len(self.set)
}

func (self *IntSet) Values() chan int64 {
	self.RLock()

	c := make(chan int64)

	go func() {
		defer self.RUnlock()

		for value := range self.set {
			c <- value
		}

		close(c)
	}()

	return c
}