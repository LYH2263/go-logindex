package stats_test

import (
	"testing"

	"github.com/LYH2263/go-logindex/internal/stats"
)

func TestTermFreqAndBM25(t *testing.T) {
	tf := stats.NewTermFreq()
	tf.AddDoc(1, map[string]int{"error": 2, "timeout": 1})
	tf.AddDoc(2, map[string]int{"error": 1})
	if tf.TF("error") != 3 || tf.DF("error") != 2 {
		t.Fatalf("tf/df %d %d", tf.TF("error"), tf.DF("error"))
	}
	ds := stats.NewDocStats()
	ds.Set(1, 3)
	ds.Set(2, 1)
	score := stats.BM25(2, 2, 2, 3, ds.AvgLength(), 1.2, 0.75)
	if score <= 0 {
		t.Fatalf("bm25 %v", score)
	}
	snap := stats.Capture(tf, ds, 5)
	if snap.NumDocs != 2 {
		t.Fatalf("snap %+v", snap)
	}
}
