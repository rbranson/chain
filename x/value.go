package x

import "reflect"

// Nil returns true if v is nil, respecting the various caveats of nullable
// types that don't "equal" nil when boxed by interface{}.
func Nil(v interface{}) bool {
	_, ok := ValueOf(v)
	return !ok
}

// ValueOf returns the reflect.Value of v if it is not nil, respecting the
// various caveats of nullable types that don't "equal" nil when boxed
// by interface{}.
func ValueOf(v interface{}) (reflect.Value, bool) {
	zero := reflect.Value{}

	if v == nil {
		return zero, false
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		if rv.IsNil() {
			return zero, false
		}
	}

	return rv, true
}
