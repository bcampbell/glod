package main

import (
	"fmt"
	"reflect"
	"sort"
)

// helper for sorting arbitrary things
type pair struct {
	key reflect.Value
	val reflect.Value
}

type sortamajig []pair

func (s sortamajig) Len() int {
	return len(s)
}

func (s sortamajig) Less(i, j int) bool {

	a := s[i].key
	b := s[j].key
	ak := a.Kind()
	bk := b.Kind()

	if ak == bk {
		switch ak {
		case reflect.String:
			return a.String() < b.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return a.Int() < b.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return a.Uint() < b.Uint()
		case reflect.Float32, reflect.Float64:
			return a.Float() < b.Float()
		}
	}
	// could potentially do more (eg string<->number coercion) but we're firmly
	// in the realm of steadily-diminishing returns by this point.
	return false
}

func (s sortamajig) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func evalField(obj reflect.Value, field string) (reflect.Value, error) {
	// TODO: should handle structs, but unlikely to see them here in glod.
	// (but slices/arrays with numeric field might be worth adding...)
	// TODO: should handle nested syntax for field "foo.wibble.name"
	if obj.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("%s doesn't have field %s (only works on maps)", obj.Kind(), field)
	}
	return indirect(obj.MapIndex(reflect.ValueOf(field))), nil
}

func indirect(v reflect.Value) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Ptr, reflect.Interface:
			v = v.Elem()
		default:
			return v
		}

	}
	// never get here
}

// sortStuff sorts an arbitrary sequence of elements
// optional args:
//   sortField  string - name of field to sort by (eg "date")
//                       if missing, maps are sorted by key and
//                       slices/arrays sorted by value
//   sortOrder  string - must be "asc" (default) or "desc"
func sortStuff(stuff interface{}, args ...interface{}) (interface{}, error) {

	// process the optional arguments
	sortField := ""
	sortAsc := true
	if len(args) > 0 {
		if fld, ok := args[0].(string); ok {
			sortField = fld
		} else {
			return nil, fmt.Errorf("sort field must be a string")
		}
	}
	if len(args) > 1 {
		if ord, ok := args[1].(string); ok {
			switch ord {
			case "desc":
				sortAsc = false
			case "asc":
				sortAsc = true
			default:
				return nil, fmt.Errorf("sort order must be asc or desc (got %s)", ord)
			}
		} else {
			return nil, fmt.Errorf("sort order must be a string")
		}
	}

	// TODO: handle indirection (pointers)
	v := reflect.ValueOf(stuff)
	kind := v.Kind()

	switch kind {
	case reflect.Map, reflect.Slice, reflect.Array:
	default:
		return nil, fmt.Errorf("sort doesn't work upon %s", kind.String())
	}

	// build up a sortamajig with sortkey-value pairs for sorting
	// TODO: discard pairs with zero keys?
	//       not so ideal, but probably more useful in the context of glod...
	//		 (ie sort by "date", also filters out pages without dates)
	s := make(sortamajig, v.Len())
	switch kind {
	case reflect.Map:
		if sortField == "" {
			// no sort field - sort by key
			for i, mapKey := range v.MapKeys() {
				s[i] = pair{key: mapKey, val: v.MapIndex(mapKey)}
			}
		} else {
			// sort by specific field in value
			for i, mapKey := range v.MapKeys() {
				val := v.MapIndex(mapKey)
				key, err := evalField(val, sortField)
				if err != nil {
					return nil, err
				}
				s[i] = pair{key: key, val: val}
			}
		}

	case reflect.Slice, reflect.Array:
		if sortField == "" {
			// no sort field - sort value
			for i := 0; i < v.Len(); i++ {
				val := v.Index(i)
				s[i] = pair{key: val, val: val}
			}
		} else {
			// sort by specific field in value
			for i := 0; i < v.Len(); i++ {
				val := v.Index(i)
				key, err := evalField(val, sortField)
				if err != nil {
					return nil, err
				}
				s[i] = pair{key: key, val: val}
			}
		}
	}

	if sortAsc {
		sort.Sort(s)
	} else {
		sort.Sort(sort.Reverse(s))
	}
	/*
		for i := 0; i < len(s); i++ {
			fmt.Printf("%d': %v\n", i, s[i].key)
		}
	*/

	// all done. now build up slice of sorted values to return
	sliceType := reflect.SliceOf(v.Type().Elem())
	out := reflect.MakeSlice(sliceType, len(s), len(s))
	for i, item := range s {
		out.Index(i).Set(item.val)
	}
	return out.Interface(), nil
}
