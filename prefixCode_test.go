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
	5) print itself as prefcode.
	6) swap labels on two leaves (SwapPermAtKeys)
	7) change permutation of labels (ApplyPerm)
	8) print perm to string (natural order leaves as ints are index for labels.)
	9) Meet of two prefix codes
	10)join of two prefix codes

	TODO (23-05-2020):
	1) Better tests around ExpandAt and DFS string creation
	2) Testing output as DFS string.
	3) Test ability to say which part of the code is a prefix of a long enough entry.
*/
func Test(t *testing.T) {

	assertCorrectMessage := func(t *testing.T, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}

	// String gives correctly formatted output
	t.Run("ExpandAt deeper than prefcode unicodealphastring.", func(t *testing.T) {
		baseCodeToExpand := NewPrefCodeAlphaString("日本語")
		baseCodeToExpand.ExpandAt("本")
		assertCorrectMessage(t, strconv.Itoa(baseCodeToExpand.Size()), "5")
	})

	// ExpandAt("here") for "here" having an element of code as a proper prefix
	// (expands to minimal prefix code with dangling caret at "here").
	t.Run("ExpandAt deeper than prefcode no init alphabet.", func(t *testing.T) {
		baseCodeToExpand := NewPrefCode()
		baseCodeToExpand.ExpandAt("1001")
		assertCorrectMessage(t, baseCodeToExpand.String(), "[0 0], [1000 1], [10010 2], [10011 3], [101 4], [11 5]")
	})

	// ExpandAt("here") for "here" properly in the body of tree.
	t.Run("ExpandAt shallower than prefcode.", func(t *testing.T) {
		myRunes := StringToRuneSlice("01")
		baseCodeToExpand := NewPrefCodeAlphaRunes(myRunes)
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

	// Join Test.
	t.Run("Join a right vine tree to a left vine tree.", func(t *testing.T) {
		myRunes := StringToRuneSlice("01")
		pcFirst := NewPrefCodeAlphaRunes(myRunes)
		pcFirst.ExpandAt("0001")
		pcSecond := NewPrefCodeAlphaRunes(myRunes)
		pcSecond.ExpandAt("1101")
		pcJoin := pcFirst.Join(pcSecond)
		got := pcJoin.String()
		want := "[0000 0], [00010 1], [00011 2], [001 3], [01 4], [10 5], [1100 6], [11010 7], [11011 8], [111 9]"
		assertCorrectMessage(t, got, want)
	})

	// Meet Test
	t.Run("Meet a right vine tree to anothertree.", func(t *testing.T) {
		myRunes := StringToRuneSlice("01")
		pcFirst := NewPrefCodeAlphaRunes(myRunes)
		pcFirst.ExpandAt("1101")
		pcSecond := NewPrefCodeAlphaRunes(myRunes)
		pcSecond.ExpandAt("1111")
		pcMeet := pcFirst.Meet(pcSecond)
		got := pcMeet.String()
		want := "[0 0], [10 1], [110 2], [111 3]"
		assertCorrectMessage(t, got, want)
	})

	// SwapPermAtKeys swaps label values at two prefixcode keys.
	t.Run("Checking SwapPermAtKeys.",
		func(t *testing.T) {
			baseCode := NewPrefCode()
			baseCode.ExpandAt("1001")

			baseCode.SwapPermAtKeys("0", "11")

			got := PermToString(baseCode.Permutation())
			want := "[0 5], [1 1], [2 2], [3 3], [4 4], [5 0]"

			assertCorrectMessage(t, got, want)
		})

	// ApplyPerm applies a given permutation to the integer key labels.
	t.Run("Checking ApplyPerm.",
		func(t *testing.T) {
			baseCode := NewPrefCode()
			baseCode.ExpandAt("1001")
			baseCode.SwapPermAtKeys("0", "11")
			baseCode.ApplyPerm(map[int]int{0: 3, 1: 5, 2: 1, 3: 0, 4: 2, 5: 4})

			got := PermToString(baseCode.Permutation())
			want := "[0 4], [1 5], [2 1], [3 0], [4 2], [5 3]"
			assertCorrectMessage(t, got, want)
		})

	// Checks whether we can use := to copy prefix codes.
	t.Run("Checking assignment.",
		func(t *testing.T) {
			baseCode := NewPrefCode()
			baseCode.ExpandAt("1001")
			baseCode.SwapPermAtKeys("0", "11")
			baseCode.ApplyPerm(map[int]int{0: 3, 1: 5, 2: 1, 3: 0, 4: 2, 5: 4})

			var targetCode PrefCode

			targetCode = baseCode

			got := targetCode.String()
			//fmt.Println("got:  " + got)
			want := baseCode.String()
			//fmt.Println("want: " + want)
			assertCorrectMessage(t, got, want)
		})

}
