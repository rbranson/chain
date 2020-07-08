package assert

/*

Adapted from github.com/benbjohnson/testing:

The MIT License (MIT)

Copyright (c) 2014 Ben Johnson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/rbranson/chain/x"
)

func baseAssert(t *testing.T, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(2)
		t.Fatalf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
	}
}

// Assert fails the test if the condition is false.
func Assert(t *testing.T, condition bool, msg string, v ...interface{}) {
	baseAssert(t, condition, msg, v...)
}

// Ok fails the test if an err is not nil.
func Ok(t *testing.T, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(t *testing.T, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		diff := cmp.Diff(exp, act, cmpopts.IgnoreUnexported())
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\033[31m%s:%d:\n\ndiff:\n%s\n\n\033[39m\n\n", filepath.Base(file), line, diff)
	}
}

// True fails the test if the condition is false
func True(t *testing.T, condition bool) {
	baseAssert(t, condition, "not true")
}

// False fails the test if the condition is true
func False(t *testing.T, condition bool) {
	baseAssert(t, !condition, "not false")
}

// Implements fails the test if v doesn't implement an interface type from exp
func Implements(t *testing.T, exp interface{}, v interface{}) {
	expExample, err := x.MakeTypeExample(exp)
	if err != nil {
		t.Fatalf("expected interface type error: %v", err)
	}

	baseAssert(t, !x.Nil(v), "passed value %v must not be nil", v)
	baseAssert(
		t,
		expExample.AssignableFrom(reflect.TypeOf(v)),
		"%v does not implement %v",
		reflect.TypeOf(v),
		expExample.Type(),
	)
}

// Panics fails the test if running f doesn't panic with msg
func Panics(t *testing.T, msg string, f func()) {
	defer func() {
		r := recover()
		baseAssert(t, r == msg, "expected panic with %v, got %v", msg, r)
	}()
	f()
}
