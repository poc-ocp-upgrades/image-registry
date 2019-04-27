package counter

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCounter(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	c := New()
	mustEqual := func(expected M) {
		if diff := c.Diff(expected); diff != nil {
			t.Fatalf("%v is expected to be equal to %v; got difference: %q", c.Values(), expected, diff)
		}
	}
	mustDiff := func(m M, want string) {
		diff := c.Diff(m)
		if got := fmt.Sprintf("%q", diff); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
	c.Add(100, 1)
	c.Add(200, 2)
	c.Add(300, 3)
	mustEqual(M{100: 1, 200: 2, 300: 3})
	c.Add(200, -2)
	mustEqual(M{100: 1, 200: 0, 300: 3})
	mustEqual(M{100: 1, 300: 3})
	mustEqual(M{100: 1, 300: 3, 400: 0})
	if values, expected := c.Values(), (M{100: 1, 200: 0, 300: 3}); !reflect.DeepEqual(values, expected) {
		t.Fatalf("got %v, want %v", values, expected)
	}
	mustDiff(M{100: 1}, `["300: got 3, want 0"]`)
	mustDiff(M{100: 1, 300: 3, 400: 4}, `["400: got 0, want 4"]`)
}
