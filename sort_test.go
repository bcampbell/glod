package main

import (
	"reflect"
	"testing"
)

func TestSortStuff(t *testing.T) {

	type objtype map[string]interface{}
	obj1 := objtype{"num": 1, "name": "foxtrot"}
	obj2 := objtype{"num": 2, "name": "echo"}
	obj3 := objtype{"num": 3, "name": "charlie"}
	obj4 := objtype{"num": 4, "name": "bravo"}
	obj5 := objtype{"num": 5, "name": "alpha"}

	foo := []struct {
		in     interface{}
		args   []interface{}
		expect interface{}
	}{
		// empty set
		{[]string{}, []interface{}{}, []string{}},
		// sort slice by value
		{[]string{"durian", "banana", "cherry", "apple"}, []interface{}{}, []string{"apple", "banana", "cherry", "durian"}},
		{[]string{"durian", "banana", "cherry", "apple"}, []interface{}{"", "desc"}, []string{"durian", "cherry", "banana", "apple"}},
		// sort by map key
		{map[string]string{"three": "durian", "one": "banana", "two": "cherry", "zero": "apple"}, []interface{}{}, []string{"banana", "durian", "cherry", "apple"}},
		// sort slice by map - nonsensical, so just does nothing
		{[]objtype{obj1, obj2, obj3, obj4, obj5}, []interface{}{}, []objtype{obj1, obj2, obj3, obj4, obj5}},
		// sort slice by field in map
		{[]objtype{obj4, obj2, obj1, obj3, obj5}, []interface{}{"name"}, []objtype{obj5, obj4, obj3, obj2, obj1}},
		{[]objtype{obj4, obj2, obj1, obj3, obj5}, []interface{}{"num"}, []objtype{obj1, obj2, obj3, obj4, obj5}},
	}

	for _, f := range foo {

		got, err := sortStuff(f.in, f.args...)
		if err != nil {
			t.Errorf("Expected %v but got error %s", f.expect, err)
		}
		if !reflect.DeepEqual(f.expect, got) {
			t.Errorf("Expected %v but got %v", f.expect, got)
		}
	}
}
