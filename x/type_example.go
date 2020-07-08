package x

import (
	"errors"
	"reflect"
)

var (
	// ErrIncorrectType indicates that the provided example is not of the
	// correct type family.
	ErrIncorrectType = errors.New("must be a pointer")
)

// TypeExample is useful to check if other types can be assigned to a given
// type using an example like (*MyType)(nil).
type TypeExample struct {
	t reflect.Type
}

// MakeTypeExample returns a new TypeExample given the passed example value,
// or returns an error. The passed example must be a pointer type.
func MakeTypeExample(example interface{}) (TypeExample, error) {
	zero := TypeExample{}
	exampleType := reflect.TypeOf(example)
	if exampleType.Kind() != reflect.Ptr {
		return zero, ErrIncorrectType
	}

	e := TypeExample{
		t: exampleType.Elem(),
	}
	return e, nil
}

// Type returns the element type of the example (not the pointer type).
func (e *TypeExample) Type() reflect.Type {
	return e.t
}

// AssignableFrom returns true if the other type can be assigned to this
// type example.
func (e *TypeExample) AssignableFrom(other reflect.Type) bool {
	return other.AssignableTo(e.t)
}
