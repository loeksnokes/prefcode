package prefcode

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

/*
PrefCode:
==========================
==========================

An interface built specifically for the type prefixCode.

prefixCode::

Has:
alphabet []rune:  The alphabet is ordered by natural rune ordering but listing in slice is arbitrary.

code map[string]int:  The strings are the elements of a finite
complete prefix code in dictionary order over the alphabet of
Runes (using natural rune order), the ints are a permutation of {0 ... n-1}
where there are n keys in the map.

Can:
    1) Verify it is a prefix code.
    2) Describe its cardinality.
	3) expand a caret at a string with all
		collateral consequenes (leaf or
		deeper, ignores shallow locations,
		all later points re-indexed)
	4) reduce at a string (shallow locations
	    only make an impact, all later points re-indexed)
    5) list exposed carets.
    6) print itself.
===================================================
TODO: safety checking: expandAt/reduceAt do not currently check if string is legal for alphabet.
*/

// PrefCode is interface for struct prefixCode which attempts to represent only
// complete finite prefix codes over a finite alphabet.
type PrefCode interface {
	Alphabet() []rune
	SetAlphabet([]rune)
	Equals(p *PrefCode) bool
	ReduceAt(s string)
	ExpandAt(s string)
	//	Join(p1, p2 PrefixCode) *PrefixCode
	//	Meet(p1, p2 PrefixCode) *PrefixCode
	ExposedCarets() []string
	Size() int
	String() string
}

type prefixCode struct {
	alphabet []rune
	code     map[string]int
}

// NewPrefCode returns a prefixCode as a PrefCode.  Magically sets the alphabet to be "01".
// use NewPrefCodeAlpha to instantiate a code with a different alphabet.
func NewPrefCode() PrefCode {
	var prefc prefixCode
	prefc.alphabet = []rune("01")
	prefc.code = map[string]int{"0": 0, "1": 1}
	return prefc
}

// NewPrefCodeAlpha returns a prefixCode as a PrefCode and sets alphabet of runes by input string.
func NewPrefCodeAlpha(alphaStr string) PrefCode {
	var prefc prefixCode
	prefc.alphabet = MakeAlphabet(alphaStr)
	prefc.code = make(map[string]int)
	for k, v := range prefc.alphabet {
		prefc.code[string(v)] = k
	}
	return prefc
}

//returns a ptr to a copy of the alphabet runes.
func (p prefixCode) Alphabet() []rune {
	retVal := make([]rune, len(p.alphabet))
	for k, v := range p.alphabet {
		retVal[k] = v
	}
	return retVal
}

func (p prefixCode) Size() int {
	return len(p.code)
}

func (p prefixCode) String() string {

	keys := make([]string, 0, len(p.code))
	for k := range p.code {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var build string
	for _, k := range keys {
		build += "[" + k + " " + strconv.Itoa(p.code[k]) + "]" + ", "
	}

	trimLen := len(build) - 2
	if trimLen < 0 {
		trimLen = 0
	}
	return strings.TrimSuffix(build, ", ")
}

func (p prefixCode) SetAlphabet(a []rune) {
	p.alphabet = make([]rune, len(a))
	copy(p.alphabet, a)
}

func (p prefixCode) Equals(q *PrefCode) bool {
	return p.String() == (*q).String()
}

//ReduceAt replaces tree dangling at s with
//just s and updates values of the PrefixCode.
func (p prefixCode) ReduceAt(s string) {
	foundCount := 0
	foundKey := false
	firstFoundix := len(p.code)

	//	fmt.Println("ReduceAt(" + s + ") start: ")
	//	fmt.Println(p.String())

	for k, v := range p.code {
		if strings.HasPrefix(k, s) {
			if !foundKey {
				foundKey = true
			}
			if v < firstFoundix {
				firstFoundix = v
			}
			foundCount++
			delete(p.code, k)
		}
	}
	if foundKey {
		p.code[s] = firstFoundix
		for k, v := range p.code {
			if v > firstFoundix {
				p.code[k] = v + 1 - foundCount
			}
		}
	}
	//	fmt.Println("ReduceAt(" + s + ") result: ")
	//	fmt.Println(p.String())

}

//expandAt adds a dangling tree to the prefix r of t
//that resides in the PrefixCode, if such exists.  It
//adds the minimal tree rooted at r so that the result
//contains t as a member of the code.
func (p prefixCode) ExpandAt(s string) {
	//how much do we need to add to expand to s from p.
	var lengthDiff int
	var prefix string
	var indexP int
	var numberNewCodes int
	var buildSpine []rune
	var toAppend []string
	//fmt.Println("ExpandAt(" + s + ") start: ")
	//fmt.Println(p.String())

	// this is all made more complicated as our string
	// has runes, not chars, so slices index poorly (by my current reading)
	for k, v := range p.code {
		if strings.HasPrefix(s, k) { //if s has k as a prefix ...
			indexP = v
			prefix = k
			lengthDiff = len(s) - len(k)
			numberNewCodes = lengthDiff*(len(p.alphabet)-1) + len(p.alphabet)
			buildSpine = []rune(strings.TrimPrefix(s, k))
			//index counting strings we are creating to add to the PrefCode
			ii := 0
			//container for new strings
			toAppend = make([]string, numberNewCodes)
			if lengthDiff > 0 {
				var v rune
				for jj := 0; jj < len(buildSpine); jj++ {
					v = buildSpine[jj]
					for _, r := range p.alphabet {
						if r != v {
							// maybe incorrect conversion back and forth for runes
							// slices are structured over bytes.
							toAppend[ii] = string(buildSpine[:jj]) + string(r)
							ii++
						}
					}
				}
			}
			// full alphabet expansion one rune beyond s
			for _, r := range p.alphabet {
				toAppend[ii] = string(buildSpine[:]) + string(r)
				ii++
			}
		}
	}

	if nil != toAppend {
		sort.Strings(toAppend)

		// delete k from p.code.
		// then reindex the later keys by adding numberNewCodes-1
		// (we are adding numberNewCodes new strings but deleted one)
		// then insert the new codes to the prefixCode
		delete(p.code, prefix)
		for lateKey, v := range p.code {
			if v > indexP {
				p.code[lateKey] = v + numberNewCodes - 1
			}
		}
		for jj, v := range toAppend {
			p.code[prefix+v] = indexP + jj
		}
	}
	//fmt.Println("ExpandAt(" + s + ") result: ")
	//fmt.Println(p.String())
}

func (p prefixCode) ExposedCarets() (caretRoots []string) {
	mset := make(map[string]string) // New empty multiset
	var prefLen int
	var thisString string

	//	fmt.Println("Searching for exposed carets in the prefix code: ")
	//	fmt.Println(p.String())

	for k, v := range p.code {
		prefLen = len(k)
		if prefLen > 0 {
			thisString = trimLastChar(k)
			//			fmt.Println("key: " + k + "  trimmed: " + thisString)
			mset[thisString] = mset[thisString] + strconv.Itoa(v)
		}
	}
	//	fmt.Println("the shortened words are: ")
	//	fmt.Println(mset)
	alphaSize := len(p.alphabet)
	for k, v := range mset {
		if len(v) == alphaSize {
			caretRoots = append(caretRoots, k)
			//			fmt.Println("Added " + k + "to caretRoots.")
		}
	}
	sort.Strings(caretRoots)
	return
}

/*
func (p PrefixCode) Join(p1, p2 PrefixCode) *PrefixCode{

}

func (p PrefixCode) Meet(p1, p2 PrefixCode) *PrefixCode{

}


func (p PrefixCode) Size() int{

}
*/

// StringToRuneSlice converts a string to a slice of runes.
func StringToRuneSlice(s string) []rune {
	var r []rune
	for _, runeValue := range s {
		r = append(r, runeValue)
	}
	return r
}

// trimLastChar consumes the last digit of string.
func trimLastChar(s string) string {
	r, size := utf8.DecodeLastRuneInString(s)
	if r == utf8.RuneError && (size == 0 || size == 1) {
		size = 0
	}
	return s[:len(s)-size]
}

// SortStringByCharacter sorts a string of runes
func SortStringByCharacter(s string) string {
	r := StringToRuneSlice(s)
	sort.Slice(r, func(i, j int) bool {
		return r[i] < r[j]
	})
	return string(r)
}

// MakeAlphabet turns a string of runes into a sorted slice of runes without duplicates.
func MakeAlphabet(s string) []rune {
	if len(s) == 0 {
		return nil
	}

	r := StringToRuneSlice(s)
	var a []rune

	sort.Slice(r, func(i, j int) bool {
		return r[i] < r[j]
	})

	//r is not empty or we would have returned on empty string.
	a = append(a, r[0])
	curIndex := 0
	for _, v := range r {
		if v != a[curIndex] {
			a = append(a, v)
			curIndex++
		}
	}

	return a
}

// dictOrder returns an integer comparing the two byte slices,
// lexicographically.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b
func dictOrder(a, b []rune) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		switch {
		case a[i] > b[i]:
			return 1
		case a[i] < b[i]:
			return -1
		}
	}
	switch {
	case len(a) > len(b):
		return 1
	case len(a) < len(b):
		return -1
	}
	return 0
}
