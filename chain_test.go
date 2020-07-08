package chain_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rbranson/chain"
	"github.com/rbranson/chain/iface"
	"github.com/rbranson/chain/internal/assert"
	"github.com/rbranson/chain/x"
)

type nonunwrappable struct{}

type unwrappable struct {
	wrapped chain.Holder
}

func (u *unwrappable) Unwrap() (interface{}, bool) {
	return u.wrapped.Get()
}

type unwrappable2 struct {
	wrapped chain.Holder
}

func (u *unwrappable2) Unwrap() (interface{}, bool) {
	return u.wrapped.Get()
}

func TestUnwrapNonUnwrappable(t *testing.T) {
	u, ok := chain.Unwrap(&nonunwrappable{})
	assert.False(t, ok)
	assert.Equals(t, nil, u)
}

func TestUnwrapUnwrappable(t *testing.T) {
	inner := &struct{}{}
	w := &unwrappable{wrapped: chain.Hold(inner)}
	assert.Implements(t, (*iface.Unwrap)(nil), w)

	u, ok := chain.Unwrap(w)
	assert.True(t, ok)
	assert.Equals(t, inner, u)
}

func TestUnwrapRecursive(t *testing.T) {
	depth := 10
	var w iface.Unwrap = &unwrappable{}
	for i := 0; i < depth; i++ {
		w = &unwrappable{wrapped: chain.Hold(w)}
	}

	cnt := 0
	var i interface{} = w
	for {
		var ok bool
		i, ok = chain.Unwrap(i)
		if !ok {
			break
		}
		cnt++
	}

	assert.Equals(t, depth, cnt)
}

type isMatcher struct {
	to interface{}
}

func (m *isMatcher) Is(other interface{}) bool {
	return reflect.DeepEqual(other, m.to)
}

func TestIs(t *testing.T) {
	var nilSlice []struct{}

	assert.True(t, chain.Is(nil, nil))
	assert.False(t, chain.Is(nil, nilSlice))
	assert.True(t, chain.Is(nilSlice, nilSlice))

	w1 := &unwrappable{}
	w2 := &unwrappable{wrapped: chain.Hold(&unwrappable2{})}

	assert.True(t, chain.Is(w1, w1))
	assert.True(t, chain.Is(w1, &unwrappable{}))
	assert.False(t, chain.Is(w1, &unwrappable2{}))
	assert.False(t, chain.Is(w1, &struct{}{}))

	assert.True(t, chain.Is(w2, w2))
	assert.False(t, chain.Is(w2, &unwrappable{}))
	assert.True(t, chain.Is(w2, &unwrappable2{}))
	assert.False(t, chain.Is(w2, &struct{}{}))

	m1 := &isMatcher{to: &unwrappable{}}
	assert.True(t, chain.Is(m1, &unwrappable{}))
	assert.False(t, chain.Is(m1, &unwrappable2{}))

	m2 := &isMatcher{to: &unwrappable2{}}
	assert.False(t, chain.Is(m2, &unwrappable{}))
	assert.True(t, chain.Is(m2, &unwrappable2{}))
}

type asMatcher struct {
	to interface{}
}

func (m *asMatcher) As(other interface{}) bool {
	// intentionally doesn't set anything here
	return reflect.DeepEqual(other, m.to)
}

func TestAs(t *testing.T) {
	assert.Panics(t, "chain: target must not be nil", func() {
		chain.As(nil, nil)
	})

	assert.Panics(t, "chain: target must be a pointer", func() {
		chain.As(nil, "")
	})

	// basic type coalescing is fair game
	var ps string
	assert.True(t, chain.As("abc", &ps))
	assert.Equals(t, "abc", ps)

	u2 := &unwrappable2{}

	w1 := &unwrappable{}
	w2 := &unwrappable{wrapped: chain.Hold(u2)}
	w3 := &unwrappable{wrapped: chain.Hold(w2)}

	hs := "hello"
	am := &asMatcher{to: &hs}
	w4 := &unwrappable{wrapped: chain.Hold(am)}

	suite := []struct {
		w                    interface{}
		expectAsUnwrappable  interface{}
		expectAsUnwrappable2 interface{}
		expectOther          interface{}
	}{
		{
			w:                    w1,
			expectAsUnwrappable:  w1,
			expectAsUnwrappable2: nil,
		},
		{
			w:                    w2,
			expectAsUnwrappable:  w2,
			expectAsUnwrappable2: u2,
		},
		{
			w:                    w3,
			expectAsUnwrappable:  w3,
			expectAsUnwrappable2: u2,
		},
		{
			w:                    w4,
			expectAsUnwrappable:  w4,
			expectAsUnwrappable2: nil,
			expectOther:          &hs,
		},
	}

	for i, tc := range suite {
		t.Run(fmt.Sprintf("tc%d", i), func(t *testing.T) {
			assert.True(t, chain.As(tc.w, &tc.w))

			eval := func(exp interface{}, target interface{}) {
				ok := chain.As(tc.w, target)
				elem := reflect.ValueOf(target).Elem().Interface()

				if x.Nil(exp) {
					assert.False(t, ok)
					assert.True(t, x.Nil(elem))
					return
				}

				assert.True(t, ok)
				assert.Equals(t, exp, elem)
			}

			var uw *unwrappable
			eval(tc.expectAsUnwrappable, &uw)

			var uw2 *unwrappable2
			eval(tc.expectAsUnwrappable2, &uw2)

			if !x.Nil(tc.expectOther) {
				assert.True(t, chain.As(tc.w, tc.expectOther))
			}
		})
	}

	// check false case for As interface assertion
	hs2 := "olleh"
	assert.True(t, hs != hs2)
	assert.False(t, chain.As(w4, &hs2))
	assert.Equals(t, "olleh", hs2)
}

func TestBuild(t *testing.T) {
	assert.Panics(t, "chain: Build called with zero arguments", func() {
		chain.Build()
	})

	foo := &struct{}{}
	assert.True(t, chain.Build(foo) == foo)

	ch1 := chain.Build("a", "b", "c")
	assert.True(t, chain.Is(ch1, "a"))
	assert.True(t, chain.Is(ch1, "b"))
	assert.True(t, chain.Is(ch1, "c"))
	assert.False(t, chain.Is(ch1, "d"))

	var s1 string
	assert.True(t, chain.As(ch1, &s1))
	assert.Equals(t, "c", s1)

	ch1b, ok := chain.Unwrap(ch1)
	assert.True(t, ok)
	assert.True(t, chain.As(ch1b, &s1))
	assert.Equals(t, "b", s1)

	var ch1l1 chain.Link
	assert.False(t, chain.As(ch1, &ch1l1))

	ch2l1 := &chain.Link{}
	ch2l2 := &chain.Link{}
	ch2l3 := &chain.Link{}
	ch2l1.Set("1")
	ch2l2.Set("2")
	ch2l3.Set("3")

	ch2 := chain.Build(ch2l1, ch2l2, ch2l3)
	assert.True(t, chain.Is(ch2, "1"))
	assert.True(t, chain.Is(ch2, "2"))
	assert.True(t, chain.Is(ch2, "3"))
	assert.False(t, chain.Is(ch2, "4"))

	var ch2Link *chain.Link
	var ch2Str string
	var ch2Int int
	assert.True(t, chain.As(ch2, &ch2Link))
	assert.True(t, chain.As(ch2, &ch2Str))
	assert.Equals(t, ch2Str, "3")
	assert.True(t, ch2Link.Is("3"))
	assert.False(t, chain.As(ch2, &ch2Int))
}
