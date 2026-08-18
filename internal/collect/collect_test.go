package collect_test

import (
	"testing"

	"github.com/LYH2263/go-logindex/internal/collect"
	"github.com/LYH2263/go-logindex/internal/posting"
)

func TestCollectorScoring(t *testing.T) {
	list := posting.New(3, 1, 2)
	col := collect.Collector{
		Scorer: collect.TFMapScorer{1: 1, 2: 5, 3: 2},
		Limit:  10,
	}
	hits := col.Collect(list)
	if len(hits) != 3 || hits[0].DocID != 2 {
		t.Fatalf("%+v", hits)
	}
	ids := collect.DocIDs(collect.SortByDocID(hits))
	if ids[0] != 1 {
		t.Fatalf("%v", ids)
	}
}
