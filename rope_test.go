package rope

import "testing"

func TestRope(t *testing.T) {
	r := New("foo!")
	for i := 0; i < 20; i++ {
		r = Concat(r, r)
	}
	if r.Len() != 4194304 {
		t.Fatalf("expected length 4194304, got %d", r.Len())
	}

	for _, c := range []struct {
		i, j int
		s    string
	}{
		{4, 8, "foo!"},
		{7, 13, "!foo!f"},
		{1024, 1028, "foo!"},
		{1027, 1034, "!foo!fo"},
	} {
		if got := r.Slice(c.i, c.j).String(); got != c.s {
			t.Errorf("expected %q, got %q", c.s, got)
		}
	}
}

func TestConcat(t *testing.T) {
	foo := "aaaaaaay62196h689161aaaaaaFFFFaaaaaaaaaa"
	bar := "bbbbbbbbbbb32632bbbbbbbbbbbbbbbbb" + foo
	baz := "cccccccc@@#@#H$H$H@$$@ccccccc" + foo + bar
	quux := bar + baz + foo + "ddd"
	r1 := Concat(New(foo), New(bar), New(baz), New(quux))
	r2 := Concat(New(foo), Concat(New(bar), New(baz), New(quux)))
	r3 := Concat(Concat(New(foo), New(bar), New(baz)), New(quux))

	for _, r := range []Rope{r1, r2, r3} {
		if r.String() != foo+bar+baz+quux {
			t.Errorf("got %q", r.String())
		}
	}
}

func TestConcatOptim(t *testing.T) {
	long := string(make([]byte, 10*small))
	short := string(make([]byte, small/2-1))
	r1 := Concat(Concat(New(long), New(short)), New(short))
	r2 := Concat(New(long), Concat(New(short), New(short)))

	if r1.Len() != 11*small-2 {
		t.Fatal()
	}
	if r2.Len() != 11*small-2 {
		t.Fatal()
	}
}

var strs = []string{
	"foo", "bar", "baz", "quux", "supercalifragilisticexpialidocious"}

func benchmarkConcat() Rope {
	r := New("bla")
	for j := 0; j < 20; j++ {
		r = Concat(r, New(strs[j%len(strs)]))
	}
	for j := 0; j < 20; j++ {
		r = Concat(r, New(""), New("x"), r)
	}
	return r
}

func BenchmarkConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkConcat()
	}
}

func BenchmarkSlice(b *testing.B) {
	r := benchmarkConcat()
	for i := 0; i < b.N; i++ {
		r.Slice(521, 1261)
		r.Slice(0, r.Len())
		r.Slice(26137, 131373)
	}
}
