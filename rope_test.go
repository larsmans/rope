package rope

import (
	"math/rand"
	"testing"
)

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

	r = Concat(r.Slice(1, 7), r.Slice(4, 16))
	if r.Len() != 18 {
		t.Errorf("expected length 18, got %d", r.Len())
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
	"foo", "bar", "baz", "quux", "supercalifragilisticexpialidocious",
	string(make([]byte, small+161)), "", "1"}

func benchmarkConcat(rng *rand.Rand) Rope {
	r := New("bla")
	for j := 0; j < 20; j++ {
		r = Concat(r, New(strs[rng.Intn(len(strs))]))
		r = Concat(New(strs[rng.Intn(len(strs))]), r)
	}
	for j := 0; j < 20; j++ {
		r = Concat(r, New(""), r, New("x"), New("y"))
	}
	return r
}

func BenchmarkConcat(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < b.N; i++ {
		benchmarkConcat(rng)
	}
}

func BenchmarkSlice(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	r := benchmarkConcat(rng)
	check := func(from, to int) {
		if n := r.Slice(from, to).Len(); n != to-from {
			b.Fatal("expected %d, got %d", to-from, n)
		}
	}

	for i := 0; i < b.N; i++ {
		check(0, 1)
		check(r.Len()-9, r.Len())
		check(521, 1261)
		check(0, r.Len())
		check(26137, 131373)
		check(1, r.Len()-1)
		check(19248, 30000)
		check(0, r.Len()-1)
	}
}
