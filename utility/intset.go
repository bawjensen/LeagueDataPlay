package utility

import (
	"fmt"
	"bytes"
	// . "github.com/bawjensen/dataplay/constants"
)

type IntSet struct {
	set map[int]bool
}

func NewIntSet(initElems ...int) (set IntSet) {
	set = IntSet{make(map[int]bool)}
	for _, elem := range initElems {
		set.set[elem] = true
	}
	return 
}

func (set IntSet) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("IntSet [ ")
	for key := range set.Values() {
		buffer.WriteString(fmt.Sprint(key, " "))
	}
	buffer.WriteString("]")
	return buffer.String()
}

func (set *IntSet) Add(elems ...int) {
	for _, elem := range elems {
		set.set[elem] = true
	}
}

func (set *IntSet) Union(other *IntSet) {
	for key := range other.Values() {
		set.Add(key)
	}
}

func (set *IntSet) Size() int {
	return len(set.set)
}

func (set *IntSet) Values() chan int {
	c := make(chan int)

	go func() {
		for value := range set.set {
			c <- value
		}

		close(c)
	}()

	return c
}