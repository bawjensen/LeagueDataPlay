package utility_test

import (
	// "fmt"
	"testing"
	// "time"
	"strings"

	. "github.com/bawjensen/dataplay/utility"
)

func TestNewIntSet(t *testing.T) {
	intSet := NewIntSet()

	if size := intSet.Size(); size != 0 {
		t.Errorf("Empty IntSet size should be 0: %d", size)
	}


	popIntSet := NewIntSet(1, 2, 3)

	if size := popIntSet.Size(); size != 3 {
		t.Errorf("Prepopulated with unique elements IntSet size should be 3: %d", size)
	}


	dupIntSet := NewIntSet(1, 1, 1, 1, 2, 3)

	if size := dupIntSet.Size(); size != 3 {
		t.Errorf("Prepopulated with duplicated elements IntSet size should be 3: %d", size)
	}
}

func TestString(t *testing.T) {
	intSet := NewIntSet(1)

	result := intSet.String()

	if !strings.Contains(result, "1") {
		t.Errorf("Stringed version of an IntSet with the element '1' should contain the substring '1': %s", result)
	}
}

func TestAdd(t *testing.T) {
	intSet := NewIntSet()

	for i := 0; i < 10; i++ {
		intSet.Add(int64(i))
	}

	if size := intSet.Size(); size != 10 {
		t.Errorf("Empty IntSet with 10 unique elements Add()'ed should have size 10: %d", size)
	}


	dupIntSet := NewIntSet()

	for i := 0; i < 10; i++ {
		dupIntSet.Add(0)
	}

	if size := dupIntSet.Size(); size != 1 {
		t.Errorf("Empty IntSet with 10 duplicated elements Add()'ed should have size 1: %d", size)
	}
}

func TestRemove(t *testing.T) {
	intSet := NewIntSet()

	intSet.Remove(1)
}

func TestHas(t *testing.T) {
	
}

func TestUnion(t *testing.T) {
	
}

func TestUnionWithout(t *testing.T) {
	
}

func TestIntersectInverse(t *testing.T) {
	
}

func TestSize(t *testing.T) {
	
}

func TestValues(t *testing.T) {
	
}
