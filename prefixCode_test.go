package prefcode

import (
	"strconv"
	"strings"
	"testing"
)

/* PrefixCode can (23-05-2020), all tested:
    1) Describe its cardinality.
    2) expand at a location.
    3) reduce at a location.
    4) list exposed carets.
	5) print itself.

	TODO (23-05-2020):
	1) Meet and Join.
	2) say which part of the code is a prefix of a long enough entry.
	3) list integer values across an exposed caret in alphabet order of children.

	Currently, a prefix code must have at least one child for each letter of
	alphabet: the empty string is not a prefix code, and this should probably
	be changed.
*/
func Test(t *testing.T) {

	assertCorrectMessage := func(t *testing.T, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}

	// String gives correctly formatted output
	t.Run("ExpandAt deeper than prefcode.", func(t *testing.T) {
		baseCodeToExpand := NewPrefCodeAlpha("日本語")
		baseCodeToExpand.ExpandAt("本")
		assertCorrectMessage(t, baseCodeToExpand.String(), "[日 0], [本日 1], [本本 2], [本語 3], [語 4]")
	})

	// Size gives correclty formatted output
	t.Run("ExpandAt deeper than prefcode.", func(t *testing.T) {
		baseCodeToExpand := NewPrefCodeAlpha("日本語")
		baseCodeToExpand.ExpandAt("本")
		assertCorrectMessage(t, strconv.Itoa(baseCodeToExpand.Size()), "5")
	})

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
