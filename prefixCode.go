package prefcode

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// FAILURE is a global constant for a failure.
const FAILURE = -1

/*
PrefCode An interface built specifically for the type prefixCode.

prefixCode::

Has:
alphabet []rune:  The alphabet is ordered by natural rune ordering but
	listing in slice is arbitrary.  The unicode code point 𝛆 which in
	utf-8 is "0xF0 0x9D 0x9B 0x86" is forbidden for the alphabet
	(representing the empty string for the empty prefix code)


code map[string]int:  The strings are the elements of a finite
complete prefix code in dictionary order over the alphabet of
Runes (using natural rune order), the ints are a permutation of {0 ... n-1}
where there are n keys in the map.

Can:
    1) Verify it is a prefix code.
    2) Describe its cardinality.
	3) expand a caret at a string with all
		collateral consequences (leaf or
		deeper, ignores shallow locations,
		all later points re-indexed)
	4) reduce at a string (shallow locations
	    only make an impact, all later points re-indexed)
    5) list exposed carets.
    6) print itself.
===================================================
TODO: safety checking: expandAt/reduceAt do not currently check if string is legal for alphabet.
*/

// EmptyString will be represented by the string "𝛆"
const EmptyString = "𝛆"

// PrefCode is interface for struct prefixCode which attempts to represent only
// complete finite prefix codes over a finite alphabet.
type PrefCode interface {
	Alphabet() []rune
	SetAlphabet([]rune)
	SetCode(map[string]int)
	Code() map[string]int
	Equals(PrefCode) bool
	ReduceAt(s string) bool
	ExpandAt(s string) bool
	ApplyPerm(perm map[int]int) bool
	SwapPermAtKeys(a, b string) error
	Permutation() map[int]int
	Join(PrefCode) (*prefixCode, error)
	Meet(PrefCode) (*prefixCode, error)
	ExposedCarets() []string
	LabelAtLeaf(string) int
	LeafAtLabel(int) string
	Size() int
	String() string
	GetPrefixOf(string) string
	CodeToSlice() *[]string
}

type prefixCode struct {
	alphabet []rune
	code     map[string]int
}

// NewPrefCode returns a prefixCode as a PrefCode.  Magically sets the alphabet to be "01".
// use NewPrefCodeAlpha to instantiate a code with a different alphabet.
func NewPrefCode() (*prefixCode, error) {
	return NewPrefCodeAlphaString("01")
}

// NewPrefCodeAlphaRunes returns a prefixCode as a PrefCode and sets alphabet of runes by slice.
func NewPrefCodeAlphaRunes(alpha []rune) (*prefixCode, error) {
	var prefc prefixCode

	//verify that 𝛆 is not a rune in alpha
	//TODO handle gracefully with better error handling.
	for _, v := range alpha {
		if EmptyString == string(v) {
			return &prefc, errors.New("Forbidden character `𝛆` in alphabet")
		}
	}

	if len(alpha) < 1 {
		return &prefc, errors.New("Empty Alphabet forbidden")
	}

	prefc.alphabet = alpha
	prefc.code = make(map[string]int, len(alpha))
	prefc.code[EmptyString] = 0
	return &prefc, nil
}

// NewPrefCodeAlphaString returns a prefixCode as a PrefCode and sets alphabet of runes by input string.
func NewPrefCodeAlphaString(alphaStr string) (*prefixCode, error) {
	return NewPrefCodeAlphaRunes(MakeAlphabet(alphaStr))
}

// DFSToPrefCode takes an alphabet of runes and a properly shaped DFS sequence
// for alphabet cardinality and creates the corresponding prefixcode with natural
// permutation.
// TODO: move to prefcode package.
func DFSToPrefCode(pc PrefCode, DFS string) bool {

	if nil == pc {
		fmt.Println("DFSToPrefCode called with nil  *PrefCode and DFS: " + DFS)
		return false
	}
	//fmt.Println("DFSToPrefCode with pc: " + (*pc).String() + " and DFS: " + DFS)
	alpha := pc.Alphabet()

	if !ValidDFSForPrefC(len(alpha), DFS) {
		fmt.Println("DFSToPrefCode: Failed ValidDFSForPrefC(" + strconv.Itoa(int(len(alpha)+'0')) + ", " + DFS + ")")
		//TODO better error handling.
		return false
	}
	var leaves []string

	// prep working stack of active words (might be extended)
	var stack []string
	currentWord := ""
	stack = append(stack, currentWord)
	top := 0 //index of top element

	// working alphabet in reverse order
	alphaSize := len(alpha)
	revAlpha := make([]rune, alphaSize)
	for ii := range alpha {
		revAlpha[ii] = alpha[alphaSize-1-ii]
	}

	for k, v := range DFS {
		if `1` == string(v) { //pop TopOStack and push alphaSize new strings on
			stack = stack[:top] //pop

			for _, letter := range revAlpha {
				stack = append(stack, currentWord+string(letter))
			}
			top = len(stack) - 1
			currentWord = stack[top]
			continue
		}
		if `0` == string(v) { //top of stack is a leaf.  Move it to leaves.
			leaves = append(leaves, stack[top])
			stack = stack[:top]
			top--
			if top < 0 {
				if k == len(DFS)-1 {
					break
				}
				// The tree filled its leaves prematurely:
				// poorly formatted.  Return Empty prefc
				return false
			}
			currentWord = stack[top]
		}
	}

	// cores is the set of words creted by deleting last letter of leaves.  Will
	// contan the exposed caret roots.
	cores := make(map[string]bool, len(leaves))
	for _, v := range leaves {
		if 0 < len(v) {
			v = v[:len(v)-1]
			cores[v] = true
		}
	}
	//fmt.Println("DFSToPrefCode with pc: " + (*pc).String() + " before expansions.")

	for k := range cores {
		if !pc.ExpandAt(k) {
			continue
		}
		//fmt.Println("DFSToPrefCode with pc: " + (*pc).String() + " after " + k + " expansion.")
	}
	//fmt.Println("DFSToPrefCode with pc: " + (*pc).String() + " after expansions.")

	return true
}

// ValidDFSForPrefC takes an integer (alphabet size) an a puported DFS string
// and verifies the string is well formatted.
func ValidDFSForPrefC(alSize int, DFS string) bool {
	// get number of carets
	carets := strings.Count(DFS, "1")
	leaves := strings.Count(DFS, "0")

	if leaves != (((alSize - 1) * carets) + 1) {
		return false
		//		panic("Wrong number of leaves in DFS tree "+DFS+" for alphabet size: "+string(`0`+alSize))
	}

	if !strings.HasPrefix(DFS, "1") {
		return false
	}
	tally := 1
	countLeaves := 0
	totalCount := 0

	for _, v := range DFS {
		totalCount++
		if "1" == string(v) {
			tally = tally + alSize - 1
		}
		if "0" == string(v) {
			tally--
			countLeaves++

		}
		if 0 == tally && totalCount < len(DFS) {
			// poorly formed DFS string.
			return false
		}
	}

	if countLeaves != (((alSize - 1) * carets) + 1) {
		return false
		//		panic("Wrong number of leaves in DFS tree "+DFS+" for alphabet size: "+string(`0`+alSize))
	}

	return true

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

func (p prefixCode) Permutation() (perm map[int]int) {
	perm = make(map[int]int, len(p.code))
	keys := make([]string, 0, len(p.code))
	for k := range p.code {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for ii, k := range keys {
		perm[ii] = p.code[k]
	}
	return
}

func (p prefixCode) SwapPermAtKeys(a, b string) error {
	valuea, oka := p.code[a]
	valueb, okb := p.code[b]
	if !oka || !okb {
		return errors.New("Did not find p.code[a] or p.code[b]")
	}
	p.code[a] = valueb
	p.code[b] = valuea

	//todo send some error too if a or b not found.
	return nil
}

// LabelAtLeaf returns the label at the leaf if it exists.
// If not, returns FAILURE global constant
func (p prefixCode) LabelAtLeaf(leaf string) (label int) {
	label, ok := p.code[leaf]

	if !ok {
		return FAILURE
	}
	return
}

// LeafAtLabel returns the leaf which carries the label, if the
// label is in bound, or the empty string otherwise.
func (p prefixCode) LeafAtLabel(label int) (leaf string) {
	//return empty string if label is out of bounds.
	//TODO: put in real error handling.
	if label > (p.Size()-1) || label < 0 {
		leaf = ""
		return
	}

	// find leaf with this label and return it.
	for k, v := range p.code {
		if v == label {
			leaf = k
			break
		}
	}
	return
}

// ApplyPerm applies a permutation map to the values of int
// labels carried by the prefixes
func (p prefixCode) ApplyPerm(perm map[int]int) bool {
	if len(p.code) != len(perm) {
		// TODO: add return for err that bad request was made.
		return false
	}

	//assumes (w/o testing) values of p.code are 0 -- k-1
	//for size k code, and likewise for perm.
	for k, v := range p.code {
		p.code[k] = perm[v]
	}
	return true
}

// PermToString converts a map[int]int into a string.
// Example Output: "[0 5], [1 1], [2 2], [3 3], [4 4], [5 0]"
func PermToString(permutation map[int]int) (permStr string) {

	// we reexpress the permutation as an ordered slice then print
	// to avoid map ordering weirdness.
	sortedPerm := make([]int, len(permutation))
	for k, v := range permutation {
		sortedPerm[k] = v
	}

	for ii, v := range sortedPerm {
		permStr += "[" + strconv.Itoa(ii) + " " + strconv.Itoa(v) + "]" + ", "
	}

	permStr = strings.TrimSuffix(permStr, ", ")
	return
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

	return strings.TrimSuffix(build, ", ")
}

func (p prefixCode) Code() map[string]int {
	return p.code
}

// No safety check, that the alphabet of the original prefixcode is the same as that of the new map.
func (p prefixCode) SetCode(pc map[string]int) {
	p.code = pc
}

func (p prefixCode) SetAlphabet(a []rune) {
	p.alphabet = make([]rune, len(a))
	copy(p.alphabet, a)
}

func (p prefixCode) Equals(q PrefCode) bool {
	return p.String() == q.String()
}

//ReduceAt replaces tree dangling at s with
//just s and updates values of the PrefixCode.
func (p prefixCode) ReduceAt(s string) bool {

	// Handle request to collapse whole PrefCode
	if "" == s || EmptyString == s {
		p.code = make(map[string]int, len(p.alphabet))
		p.code[EmptyString] = 0
		return true
	}

	// Now we face a normal request.
	// we look for s as shallower than some codes.  All such codes are
	// collapsed to s.  The permutation is re-indexed appropriately.
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
	if foundCount > 0 {
		return true
	}
	return false
}

//expandAt adds a dangling tree to the prefix r of t
//that resides in the PrefixCode, if such exists.  It
//adds the minimal tree rooted at r so that the result
//contains t as a member of the code.
//TODO: (07Aug2021) refactor logic so gocyclo count (see goreportcard on gitub) is reduced.  Should
//be easy as initial logic looks over-detected.
//E.g., 1 == len(p.code) && Emptystring==p.LeafAtLabel(0) is in both first tests.
func (p prefixCode) ExpandAt(s string) bool {

	// p.code is empty (contains EmptyString) and requested expansion is at root.
	if (EmptyString == s || "" == s) && 1 == len(p.code) && EmptyString == p.LeafAtLabel(0) {
		for k, v := range p.alphabet {
			p.code[string(v)] = k
		}
		delete(p.code, EmptyString)
		return true
	}

	// code is empty (contains EmptyString) and requested expansion is not at root.
	// (Implicit from last "If".)
	// Develop one level of p.code, then pretend we are just starting from the normal case,
	// but without the EmptyString entry in the code now.
	if 1 == len(p.code) && EmptyString == p.LeafAtLabel(0) {
		for k, v := range p.alphabet {
			p.code[string(v)] = k
		}
		delete(p.code, EmptyString)
		//do not return.  We will now pretend code was not empty and carry on.
	}

	var labelAtP int
	prefix := ""
	var lengthDiff int
	var numberNewCodes int
	buildSpine := []rune(s)

	//general handling
	var toAppend []string

	// this is all made more complicated as our string
	// has runes, not chars, so slices index poorly (by my current reading)
	// find expandAt location.
	for k, v := range p.code {
		if strings.HasPrefix(s, k) { //if s has k as a prefix ...
			labelAtP = v
			prefix = k
			lengthDiff = len(s) - len(k)
			numberNewCodes = lengthDiff*(len(p.alphabet)-1) + len(p.alphabet)
			if 0 < lengthDiff {
				buildSpine = buildSpine[len(k):] // throw away the prefix
				break
			}
			buildSpine = buildSpine[:0] //force buildSpine to be empty
			break
		}
	}
	if "" == prefix { //code is not empty but no prefix found: expansion location too shallow so do nothing.
		return false
	}

	//OK, we found a prefix and built spine for additions.
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

	if nil != toAppend {
		sort.Strings(toAppend)
		// delete k from p.code.
		// then reindex the later keys by adding numberNewCodes-1
		// (we are adding numberNewCodes new strings but deleted one)
		// then insert the new codes to the prefixCode
		delete(p.code, prefix)
		for lateKey, v := range p.code {
			if v > labelAtP {
				p.code[lateKey] = v + numberNewCodes - 1
			}
		}
		for jj, v := range toAppend {
			p.code[prefix+v] = labelAtP + jj
		}
	}
	return true
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

func (p prefixCode) GetPrefixOf(s string) string {
	for k := range p.code {
		if strings.HasPrefix(s, k) {
			return k
		}
	}
	return ""
}

// Join finds smallest prefix code so that each leaf is deeper/equal
// to leaves of both prefix codes and returns a pointer to this constructed code.
// TODO: needs testing coverage
func (p prefixCode) Join(q PrefCode) (*prefixCode, error) {
	jpc, err := NewPrefCodeAlphaRunes(p.alphabet)

	if err != nil {
		return jpc, err
	}

	expansionsP := p.ExposedCarets()

	for _, v := range expansionsP {
		jpc.ExpandAt(v)
	}

	expansionsQ := q.ExposedCarets()

	for _, v := range expansionsQ {
		jpc.ExpandAt(v)
	}
	return jpc, err
}

// Iterates from left-right through the prefx codes, choosing the shallower
// element of any comparable pair too build a new prefix code.  Replaces the first with this one.
func (p prefixCode) Meet(q PrefCode) (*prefixCode, error) {
	jpc, err := NewPrefCodeAlphaRunes(p.alphabet)

	if err != nil {
		return jpc, err
	}
	expansionsP := p.ExposedCarets()
	expansionsQ := q.ExposedCarets()

	var vRunes []rune
	var wRunes []rune

	allCommonExpansions := make(map[string]bool)
	var maxLen int
	var commonWord string

	// for each exposed caret of p and q
	// intersect words to find longest common prefix
	// the list of these can be used to expand out
	// the common tree.
	// It is OK to have several words that are extensions
	// of each other here since the expansion code will do nothing
	// when expanding shallower than the tree (it is unclear
	// if sorting to get rid of these is faster than just letting
	// a few duplicate expansions run)

	for _, v := range expansionsP {
		for _, w := range expansionsQ {
			maxLen = int(math.Min(float64(len(v)), float64(len(w))))
			vRunes = []rune(v)
			wRunes = []rune(w)
			for ii := 0; ii < maxLen; ii++ {
				if vRunes[ii] == wRunes[ii] {
					commonWord = commonWord + string(vRunes[ii])
					continue
				}
				break
			}
			if len(commonWord) > 0 {
				allCommonExpansions[commonWord] = true
				commonWord = ""
			}
		}
	}
	for k := range allCommonExpansions {
		jpc.ExpandAt(k)
	}
	return jpc, err
}

// CodeToSlice returns a * to slice consisting of the codestrings of p
func (p prefixCode) CodeToSlice() *[]string {
	codes := make([]string, len(p.code))
	for k := range p.code {
		codes = append(codes, k)
	}
	return &codes
}

// Helper functions many just found and mildly edited from standard websites.

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

// CPB authorship below:

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

// The below is a bit pointless since normal sort works fine for strings of utf8.
// it is a modified version of some code I found that worked for ascii
// and I modified it for runes.  Then I learned that < does this already for
// strings by considering codepoints.

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
