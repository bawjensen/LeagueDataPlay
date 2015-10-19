package utility

import (
	"fmt"
	"bytes"
	// . "github.com/bawjensen/dataplay/constants"
)

type IntSet struct {
	set map[int64]bool
}

func NewIntSet(initElems ...int64) (set *IntSet) {
	set = &IntSet{make(map[int64]bool)}
	for _, elem := range initElems {
		set.set[elem] = true
	}
	return set
}

func NewIntSetFromSlice(initElems []interface{}) (set *IntSet) {
	set = &IntSet{make(map[int64]bool)}
	for _, elem := range initElems {
		set.set[elem.(int64)] = true
	}
	return set
}

func (self *IntSet) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("IntSet [ ")
	for elem := range self.Values() {
		buffer.WriteString(fmt.Sprint(elem, " "))
	}
	buffer.WriteString("]")
	return buffer.String()
}

func (self *IntSet) Add(elems ...int64) {
	for _, elem := range elems {
		self.set[elem] = true
	}
}

func (self *IntSet) Has(elem int64) bool {
	return self.set[elem]
}

func (self *IntSet) Union(other *IntSet) {
	for elem := range other.Values() {
		self.Add(elem)
	}
}

func (self *IntSet) UnionWithout(other *IntSet, exclude *IntSet) {
	for elem := range other.Values() {
		if !exclude.Has(elem) {
			self.Add(elem)
		} /*else {
			fmt.Printf("Not adding %d because it was visited\n", elem)
		}*/
	}
}

func (self *IntSet) Size() int {
	return len(self.set)
}

func (self *IntSet) Values() chan int64 {
	c := make(chan int64)

	go func() {
		for value := range self.set {
			c <- value
		}

		close(c)
	}()

	return c
}