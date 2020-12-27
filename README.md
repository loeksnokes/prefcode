# prefcode
 This is a package that implements complete prefix codes in Go, over given utf8 alphabet.  Includes functionality like expanding a prefcode to a given word, 
or trimming the prefcode down at some prefix of a word in the code.  It also finds exposed carets and other optional goodies.

For example, if the alphabet is {0,1} then the set of words {0,10,11} is a complete prefix code for the dictionary order on all words:
Any long enough word has one of these three words as a prefix, and none of these words are prefixes of each other.

From the comments in prefixCode.go:

============================

```PrefCode: An interface built specifically for the type prefixCode.

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
		    deeper, ignores shallow location expansions,
		    all later points in the prefcode are re-indexed)
	  4) reduce at a string (shallow locations
	      only to make an impact, all later points in the prefcode are re-indexed)
    5) List exposed carets.
    6) Print itself.
    7) Carry a permutation of integers, over integers 0 ... k-1, in the case the code has 8 elements: one value per leaf:
        a) the code can look up which leaf is associated to which integer;
        b) the code can look up which integer is associated to which leaf.
        
**) This point (7) on permutations is not really associated to prefix codes in general, but a standard Go methodology to 
implement sets is to use maps where each key (and object of the set) has associated boolean value.  Thus, we can essentially carry 
this data free of charge, and, it useful for our intended use case (implementing R. Thompson's groups F, T, V calculations).

The code has fairly comprehensive test coverage.  It probably needs refactoring: it is one of my earliest Go projects.
    ```
