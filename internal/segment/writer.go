package segment

import (
	"fmt"
	"time"

	"github.com/LYH2263/go-logindex/internal/posting"
)

// Builder 用于从可变缓冲构造不可变段。
type Builder struct {
	termDocs map[string]map[uint64]struct{}
	live     map[uint64]struct{}
}

// NewBuilder 构造。
func NewBuilder() *Builder {
	return &Builder{
		termDocs: make(map[string]map[uint64]struct{}),
		live:     make(map[uint64]struct{}),
	}
}

// AddDoc 登记文档及词项。
func (b *Builder) AddDoc(doc uint64, terms []string) {
	if b == nil {
		return
	}
	b.live[doc] = struct{}{}
	for _, term := range terms {
		if term == "" {
			continue
		}
		m, ok := b.termDocs[term]
		if !ok {
			m = make(map[uint64]struct{})
			b.termDocs[term] = m
		}
		m[doc] = struct{}{}
	}
}

// DeleteDoc 从构建中移除。
func (b *Builder) DeleteDoc(doc uint64) {
	if b == nil {
		return
	}
	delete(b.live, doc)
	for term, m := range b.termDocs {
		delete(m, doc)
		if len(m) == 0 {
			delete(b.termDocs, term)
		}
	}
}

// DocCount 当前文档数。
func (b *Builder) DocCount() int {
	if b == nil {
		return 0
	}
	return len(b.live)
}

// Build 生成段。
func (b *Builder) Build(id ID) *Segment {
	if b == nil {
		return New(id, nil, nil, time.Now().Unix())
	}
	termDocs := make(map[string][]uint64, len(b.termDocs))
	for term, set := range b.termDocs {
		docs := make([]uint64, 0, len(set))
		for d := range set {
			if _, ok := b.live[docKey(d)]; !ok {
				continue
			}
			docs = append(docs, d)
		}
		if len(docs) == 0 {
			continue
		}
		termDocs[term] = docs
	}
	live := make([]uint64, 0, len(b.live))
	for d := range b.live {
		live = append(live, d)
	}
	return New(id, termDocs, live, time.Now().Unix())
}

func docKey(d uint64) uint64 { return d }

// MergePostings 合并多个段的同一词项 posting（调用方保证活集正确）。
func MergePostings(lists []*posting.List) *posting.List {
	return posting.UnionMany(lists)
}

// ValidateNoDup 检查同 term 同 doc 不重复（用于自检）。
func ValidateNoDup(s *Segment) error {
	if s == nil {
		return fmt.Errorf("nil segment")
	}
	for term, l := range s.AllPostings() {
		docs := l.Docs()
		for i := 1; i < len(docs); i++ {
			if docs[i] == docs[i-1] {
				return fmt.Errorf("duplicate doc %d in term %q", docs[i], term)
			}
			if docs[i] < docs[i-1] {
				return fmt.Errorf("unsorted posting for term %q", term)
			}
		}
	}
	return nil
}
