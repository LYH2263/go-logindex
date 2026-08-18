package analyze_test

import (
	"testing"

	"github.com/LYH2263/go-logindex/internal/analyze"
)

func TestDefaultAnalyzer(t *testing.T) {
	a := analyze.NewDefault()
	terms := analyze.Terms(a, "Error: Timeout connecting to the Database!")
	if len(terms) == 0 {
		t.Fatal("no terms")
	}
	// "the" "to" 应为停用词
	for _, term := range terms {
		if term == "the" || term == "to" {
			t.Fatalf("stopword leaked: %q in %v", term, terms)
		}
	}
	if !analyze.HasTerm(a, "Error Timeout", "error") {
		t.Fatal("expected error")
	}
}

func TestCamelAndNormalize(t *testing.T) {
	parts := analyze.CamelSplit("ErrorCode")
	if len(parts) < 2 {
		t.Fatalf("camel: %v", parts)
	}
	if analyze.Normalize("  AbC ") != "abc" {
		t.Fatal("normalize")
	}
}
