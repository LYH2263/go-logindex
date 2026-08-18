package posting_test

import (
	"reflect"
	"testing"

	"github.com/LYH2263/go-logindex/internal/posting"
)

func TestIntersectUnionDifference(t *testing.T) {
	a := posting.New(1, 3, 5, 7)
	b := posting.New(3, 4, 7, 9)
	if got := posting.Intersect(a, b).Docs(); !reflect.DeepEqual(got, []uint64{3, 7}) {
		t.Fatalf("intersect %v", got)
	}
	if got := posting.Union(a, b).Docs(); !reflect.DeepEqual(got, []uint64{1, 3, 4, 5, 7, 9}) {
		t.Fatalf("union %v", got)
	}
	if got := posting.Difference(a, b).Docs(); !reflect.DeepEqual(got, []uint64{1, 5}) {
		t.Fatalf("diff %v", got)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	orig := posting.New(1, 2, 10, 100, 1000, 100000)
	buf := posting.Encode(orig)
	got, err := posting.Decode(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !posting.Equal(orig, got) {
		t.Fatalf("roundtrip %v vs %v", orig.Docs(), got.Docs())
	}
}

func TestDedupOnAdd(t *testing.T) {
	l := posting.New(5, 5, 1, 3, 3)
	if !reflect.DeepEqual(l.Docs(), []uint64{1, 3, 5}) {
		t.Fatalf("%v", l.Docs())
	}
}
