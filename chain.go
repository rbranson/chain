// Package chain implements functions to manipulate chains of values.
//
// A generic implemention of the chaining logic in the errors package in the
// standard library. Who even knows what this is useful for tho?

package chain

import (
	"reflect"

	"github.com/rbranson/chain/iface"
	"github.com/rbranson/chain/x"
)

// Unwrap returns the result of calling the Unwrap method on v, if v's type
// implements iface.Unwrap. Otherwise, Unwrap returns nil and false.
func Unwrap(v interface{}) (interface{}, bool) {
	u, ok := v.(iface.Unwrap)
	if !ok {
		return nil, false
	}
	return u.Unwrap()
}

// Is reports whether any value in v's chain matches target.
//
// The chain consists of v itself followed by the sequence of values obtained
// by repeatedly calling Unwrap.
//
// A value is considered a match if it is equal to target or if it implements
// an Is(interface{}) bool such that Is(target) returns true.
//
// A value type might provide an Is method so it can be treated as equivalent
// to an existing value. For example, if MyValue defines:
//
//		func (m MyValue) Is(target interface{}) bool {
//			return target == "foo"
//    }
//
// then Is(MyValue{}, "foo") returns true.
func Is(v interface{}, target interface{}) bool {
	for {
		if x.Nil(v) && x.Nil(target) {
			return reflect.TypeOf(v) == reflect.TypeOf(target)
		}

		if isv, ok := v.(iface.Is); ok {
			if isv.Is(target) {
				return true
			}
		}

		if reflect.DeepEqual(v, target) {
			return true
		}

		var ok bool
		v, ok = Unwrap(v)
		if !ok {
			return false
		}
	}
}

// As finds the first value in v's chain that matches target, and if so, sets
// target to that value and returns true. Otherwise, it returns false.
//
// The chain consists of v itself followed by the sequence of values obtained
// by repeatedly calling Unwrap.
//
// A value matches target if its concrete value is assignable to the value
// pointed to by target, or if the value has a method As(interface{}) bool
// such that As(target) returns true. In the latter case, the As method is
// responsible for setting target.
//
// A value type might provide an As method so it can be treated as if it were
// a different value type.
//
// As panics if target is not a non-nil pointer.
func As(v interface{}, target interface{}) bool {
	targetVal, ok := x.ValueOf(target)
	if !ok {
		panic("chain: target must not be nil")
	}

	targetEx, err := x.MakeTypeExample(target)
	if err != nil {
		panic("chain: target " + err.Error())
	}

	for {
		if targetEx.AssignableFrom(reflect.TypeOf(v)) {
			targetVal.Elem().Set(reflect.ValueOf(v))
			return true
		}

		if asv, ok := v.(iface.As); ok {
			if asv.As(target) {
				return true
			}
		}

		var ok bool
		v, ok = Unwrap(v)
		if !ok {
			return false
		}
	}
}

// use an internal type to prevent people from using Link to get at it
// accidentally.
type buildLink struct {
	Link
}

func (l *buildLink) Wrap(v interface{}) bool {
	return l.Link.Wrap(v)
}

func (l *buildLink) Unwrap() (interface{}, bool) {
	return l.Link.Unwrap()
}

func (l *buildLink) Is(target interface{}) bool {
	return l.Link.Is(target)
}

func (l *buildLink) As(target interface{}) bool {
	return l.Link.As(target)
}

// Build chains together vals, returning the last element.
//
// If vals is empty, this will panic.
//
// If vals contains a single element, it is returned and no action is taken.
//
// Otherwise, a chain is built using simple rules, processed from first to last
// element. If an element implements Wrap(interface{}) bool, the Wrap method is
// called and passed the previous element. If Wrap returns true, then chaining
// continues. If the element does not implement Wrap, or Wrap returns false,
// the value is wrapped with an unspecified type and then chaining continues.
func Build(vals ...interface{}) interface{} {
	switch len(vals) {
	case 0:
		panic("chain: Build called with zero arguments")
	case 1:
		return vals[0]
	}

	src := vals[0]
	for i := 1; i < len(vals); i++ {
		dst := vals[i]

		if w, ok := dst.(iface.Wrap); ok && w.Wrap(src) {
			src = dst
			continue
		}

		link := &buildLink{}
		link.Link.Set(dst)
		if !link.Wrap(src) {
			panic("Link.Wrap should always return true")
		}

		src = link
	}

	return src
}

// Holder holds an arbitrary Value and a positive assertion that it was
// intentionaly filled.
//
// This is useful for differentiating between zero values and nils that
// were intentionally filled, or just automatically filled during
// instantiation.
//
// The reason it is included in this package is that it is a useful
// building block for implementing wrapper types for chaining.
type Holder struct {
	Value interface{}
	Ok    bool
}

// Set sets the Holder's Value to v.
func (h *Holder) Set(v interface{}) {
	h.Value = v
	h.Ok = true
}

// Get returns the Holder's Value and the "filled" assertion.
func (h *Holder) Get() (interface{}, bool) {
	if !h.Ok {
		return nil, false
	}
	return h.Value, true
}

// Hold builds a new Holder and sets it to v
func Hold(v interface{}) Holder {
	h := Holder{}
	h.Set(v)
	return h
}

// Link is a generic chainable wrapper for any value.
//
// It holds a value and wraps another value.
type Link struct {
	h Holder
	v interface{}
}

// Set sets the Link's held value to v
func (l *Link) Set(v interface{}) *Link {
	l.v = v
	return l
}

// Unwrap unwraps the wrapped value
func (l *Link) Unwrap() (interface{}, bool) {
	return l.h.Get()
}

// Wrap sets the Link's wrapped value
func (l *Link) Wrap(v interface{}) bool {
	l.h.Set(v)
	return true
}

// Is returns true if the target equals the held value
func (l *Link) Is(target interface{}) bool {
	return l.v == target
}

// As returns chain.As(v, target) where v is the held value
func (l *Link) As(target interface{}) bool {
	return As(l.v, target)
}
