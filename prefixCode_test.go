package prefcode

import (
	"strings"
	"testing"
)

/* PrefixCode can:
	1) Verify it is a prefix code is.
		a) incomparable elts.
		b) complete.
    2) Describe its cardinality.
    3) expand at a location.
    4) reduce at a location.
    5) list exposed carets.
    6) print itself.
*/
func Test(t *testing.T) {

	assertCorrectMessage := func(t *testing.T, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}

	// ExpandAt("here") for "here" having an element of code as a proper prefix
	// (expands to minimal prefix code with dangling caret at "here").
	t.Run("ExpandAt deeper than prefcode.", func(t *testing.T) {
		baseCodeToExpand := NewPrefCode()
		baseCodeToExpand.ExpandAt("1001")
		assertCorrectMessage(t, baseCodeToExpand.String(), "[0 0], [1000 1], [10010 2], [10011 3], [101 4], [11 5]")
	})

	// ExpandAt("here") for "here" properly in the body of tree.
	t.Run("ExpandAt shallower than prefcode.", func(t *testing.T) {
		baseCodeToExpand := NewPrefCode()
		baseCodeToExpand.ExpandAt("1001")
		baseCodeToExpand.ExpandAt("1")
		assertCorrectMessage(t, baseCodeToExpand.String(), "[0 0], [1000 1], [10010 2], [10011 3], [101 4], [11 5]")
	})

	// Checks ReduceAt("here") for "here" a proper prefix of some elements of the prefix code (replace all these with "here").
	t.Run("ReduceAt shallower than prefcode.", func(t *testing.T) {
		baseCode := NewPrefCode()
		baseCode.ExpandAt("1001")
		baseCode.ReduceAt("10")
		assertCorrectMessage(t, baseCode.String(), "[0 0], [10 1], [11 2]")
	})

	// Checks ReduceAt("here") for "here" having a prefix in the prefix code (do nothing).
	t.Run("ReduceAt deeper than prefcode.", func(t *testing.T) {
		baseCode := NewPrefCode()
		baseCode.ExpandAt("1001")
		baseCode.ReduceAt("11101")
		assertCorrectMessage(t, baseCode.String(), "[0 0], [1000 1], [10010 2], [10011 3], [101 4], [11 5]")
	})

	t.Run("Checking can find exposed carets.", func(t *testing.T) {
		baseCode := NewPrefCode()
		baseCode.ExpandAt("1001")
		baseCode.ExpandAt("11")
		/*
			for _, v := range baseCode.ExposedCarets() {
				fmt.Println(v)
			}
		*/
		got := strings.Join(baseCode.ExposedCarets(), " ")
		want := "1001 11"
		assertCorrectMessage(t, got, want)
	})

}
