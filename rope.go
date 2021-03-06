package rope

import (
	"bytes"
	"fmt"
	"io"
)

const small = 128

// A rope is a heavy-weight string. They are useful for manipulating very
// long strings.
//
// Like Go strings, ropes are immutable sequences of bytes.
type Rope struct {
	root tree
}

// Make a new rope with contents s.
func New(s string) Rope {
	return Rope{leaf(s)}
}

// Concatenate zero or more ropes.
func Concat(ropes ...Rope) Rope {
	switch len(ropes) {
	case 0:
		return Rope{leaf("")}
	default:
		return Rope{concat(ropes)}
	}
}

func concat(ropes []Rope) tree {
	switch n := len(ropes); n {
	case 1:
		return ropes[0].root
	case 2:
		return ropes[0].root.concat(ropes[1].root)
	default:
		n2 := n / 2
		r1, r2 := concat(ropes[:n2]), concat(ropes[n2:])
		return r1.concat(r2)
	}
}

func checkSlice(i, j int) {
	if i > j {
		panic(fmt.Sprintf("%d > %d", i, j))
	}
}

// Delete the substring at positions i through j.
func (r Rope) Delete(i, j int) Rope {
	checkSlice(i, j)
	if i < j {
		left, right := r.Slice(0, i), r.Slice(j, r.Len())
		r = Rope{left.root.concat(right.root)}
	}
	return r
}

// Returns the i'th byte in the rope.
func (r Rope) Index(i int) byte {
	return r.root.index(i)
}

// Insert ins at position i.
func (r Rope) Insert(i int, ins Rope) Rope {
	return r.Replace(i, i, ins)
}

// The length of the rope, in bytes.
func (r Rope) Len() int {
	return r.root.length()
}

// Replace positions i through j with repl.
func (r Rope) Replace(i, j int, repl Rope) Rope {
	if repl.Len() == 0 {
		return r.Delete(i, j)
	}
	checkSlice(i, j)
	return Concat(r.Slice(0, i), repl, r.Slice(j, r.Len()))
}

// Slice of a rope. Equivalent to New(r.String())[i:j].
func (r Rope) Slice(i, j int) Rope {
	checkSlice(i, j)
	if i == j {
		return Rope{leaf("")}
	}
	return Rope{r.root.slice(i, j)}
}

// Returns the contents of r as a string.
func (r Rope) String() string {
	var buf bytes.Buffer
	r.WriteTo(&buf)
	return buf.String()
}

// Write the contents of r to w. Returns the number of bytes written and
// possibly an error.
func (r Rope) WriteTo(w io.Writer) (n int64, err error) {
	return r.root.writeTo(w)
}

type tree interface {
	concat(tree) tree
	index(int) byte
	length() int
	slice(int, int) tree
	writeTo(w io.Writer) (int64, error)
}

type leaf string

func (s leaf) concat(t tree) tree {
	if s2, ok := t.(leaf); ok && len(string(s))+len(string(s2)) <= small {
		return leaf(string(s) + string(s2))
	}
	return &node{s, t, len(string(s)) + t.length()}
}

func (s leaf) index(i int) byte {
	return string(s)[i]
}

func (s leaf) length() int {
	return len(string(s))
}

func (s leaf) slice(i, j int) tree {
	return leaf(string(s)[i:j])
}

func (s leaf) writeTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(s))
	return int64(n), err
}

type node struct {
	left, right tree
	nbytes      int
}

func (n *node) concat(right tree) tree {
	var left tree
	s1, ok1 := n.right.(leaf)
	s2, ok2 := right.(leaf)
	if ok1 && ok2 && len(string(s1))+len(string(s2)) <= small {
		right = leaf(string(s1) + string(s2))
		left = n.left
	} else {
		left = n
	}
	return &node{left, right, left.length() + right.length()}
}

func (n *node) index(i int) byte {
	if i > n.nbytes {
		panic(fmt.Sprintf("%d out of bounds for length-%d rope", i, n.length))
	}
	leftlength := n.left.length()
	if i < leftlength {
		return n.left.index(i)
	}
	return n.right.index(i - leftlength)
}

func (n *node) length() int {
	return n.nbytes
}

func (n *node) slice(i, j int) tree {
	leftlen := n.left.length()
	switch {
	case j <= leftlen:
		return n.left.slice(i, j)
	case i >= leftlen:
		return n.right.slice(i-leftlen, j-leftlen)
	}

	var left, right tree
	if i == 0 {
		left = n.left
	} else {
		left = n.left.slice(i, leftlen)
	}
	if j == leftlen+n.right.length() {
		right = n.right
	} else {
		right = n.right.slice(0, j-leftlen)
	}
	return left.concat(right)
}

func (n *node) writeTo(w io.Writer) (nw int64, err error) {
	nw, err = n.left.writeTo(w)
	if err != nil {
		return
	}
	nright, err := n.right.writeTo(w)
	nw += nright
	return
}
