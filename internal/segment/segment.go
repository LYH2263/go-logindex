package segment

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/LYH2263/go-logindex/internal/posting"
)

// ID 段标识。
type ID uint64

// Meta 段元数据。
type Meta struct {
	ID        ID     `json:"id"`
	DocCount  int    `json:"doc_count"`
	MinDoc    uint64 `json:"min_doc"`
	MaxDoc    uint64 `json:"max_doc"`
	TermCount int    `json:"term_count"`
	Created   int64  `json:"created"`
}

// FailNextPersist 若非空，下一次 Persist 使用该错误（测试注入）。
var FailNextPersist error

// Segment 是不可变倒排段。
type Segment struct {
	meta      Meta
	postings  map[string]*posting.List
	liveDocs  map[uint64]struct{}
	tombstone map[uint64]struct{}
	mu        sync.RWMutex
}

// New 从词项→doc 映射构造段。
func New(id ID, termDocs map[string][]uint64, live []uint64, created int64) *Segment {
	postings := make(map[string]*posting.List, len(termDocs))
	var minDoc, maxDoc uint64
	first := true
	liveSet := make(map[uint64]struct{}, len(live))
	for _, d := range live {
		liveSet[d] = struct{}{}
		if first || d < minDoc {
			minDoc = d
			first = false
		}
		if d > maxDoc {
			maxDoc = d
		}
	}
	for term, docs := range termDocs {
		// 仅保留仍存活的文档。
		filtered := make([]uint64, 0, len(docs))
		for _, d := range docs {
			if _, ok := liveSet[d]; ok {
				filtered = append(filtered, d)
			}
		}
		if len(filtered) == 0 {
			continue
		}
		postings[term] = posting.New(filtered...)
	}
	return &Segment{
		meta: Meta{
			ID:        id,
			DocCount:  len(liveSet),
			MinDoc:    minDoc,
			MaxDoc:    maxDoc,
			TermCount: len(postings),
			Created:   created,
		},
		postings:  postings,
		liveDocs:  liveSet,
		tombstone: make(map[uint64]struct{}),
	}
}

// Meta 返回元数据拷贝。
func (s *Segment) Meta() Meta {
	if s == nil {
		return Meta{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := s.meta
	m.DocCount = len(s.liveDocs)
	return m
}

// Posting 返回词项倒排（已扣除墓碑）的拷贝。
func (s *Segment) Posting(term string) *posting.List {
	if s == nil {
		return posting.New()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.postings[term]
	if !ok {
		return posting.New()
	}
	out := l.Clone()
	if len(s.tombstone) > 0 {
		out.RemoveSet(s.tombstone)
	}
	return out
}

// Terms 返回全部词项（排序）。
func (s *Segment) Terms() []string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.postings))
	for t := range s.postings {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// LiveDocs 返回存活文档集合拷贝。
func (s *Segment) LiveDocs() map[uint64]struct{} {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[uint64]struct{}, len(s.liveDocs))
	for d := range s.liveDocs {
		out[d] = struct{}{}
	}
	return out
}

// IsLive 判断文档是否仍可见。
func (s *Segment) IsLive(doc uint64) bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, dead := s.tombstone[doc]; dead {
		return false
	}
	_, ok := s.liveDocs[doc]
	return ok
}

// MarkDeleted 添加墓碑。
func (s *Segment) MarkDeleted(doc uint64) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.liveDocs[doc]; !ok {
		return false
	}
	if _, already := s.tombstone[doc]; already {
		return false
	}
	s.tombstone[doc] = struct{}{}
	delete(s.liveDocs, doc)
	s.meta.DocCount = len(s.liveDocs)
	return true
}

// AllPostings 返回扣除墓碑后的全部倒排拷贝。
func (s *Segment) AllPostings() map[string]*posting.List {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]*posting.List, len(s.postings))
	for term, l := range s.postings {
		cp := l.Clone()
		if len(s.tombstone) > 0 {
			cp.RemoveSet(s.tombstone)
		}
		if cp.Len() == 0 {
			continue
		}
		out[term] = cp
	}
	return out
}

// DirName 段目录名。
func DirName(id ID) string {
	return fmt.Sprintf("seg-%016x", uint64(id))
}

// Persist 将段写入目录。
func (s *Segment) Persist(root string) error {
	if FailNextPersist != nil {
		// 注入的持久化失败：必须如实返回错误，否则上层会误以为落盘成功
		// 并截断 WAL，导致数据丢失。
		err := FailNextPersist
		FailNextPersist = nil
		return err
	}
	if s == nil {
		return fmt.Errorf("segment: nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	dir := filepath.Join(root, DirName(s.meta.ID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	metaBytes, err := json.MarshalIndent(s.meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0o644); err != nil {
		return err
	}
	// terms.bin: [u32 termLen][term][posting bytes]...
	var blob []byte
	terms := make([]string, 0, len(s.postings))
	for t := range s.postings {
		terms = append(terms, t)
	}
	sort.Strings(terms)
	for _, term := range terms {
		l := s.postings[term].Clone()
		if len(s.tombstone) > 0 {
			l.RemoveSet(s.tombstone)
		}
		if l.Len() == 0 {
			continue
		}
		enc := posting.Encode(l)
		var hdr [4]byte
		binary.LittleEndian.PutUint32(hdr[:], uint32(len(term)))
		blob = append(blob, hdr[:]...)
		blob = append(blob, term...)
		binary.LittleEndian.PutUint32(hdr[:], uint32(len(enc)))
		blob = append(blob, hdr[:]...)
		blob = append(blob, enc...)
	}
	if err := os.WriteFile(filepath.Join(dir, "postings.bin"), blob, 0o644); err != nil {
		return err
	}
	// live docs
	live := make([]uint64, 0, len(s.liveDocs))
	for d := range s.liveDocs {
		live = append(live, d)
	}
	sort.Slice(live, func(i, j int) bool { return live[i] < live[j] })
	lb := make([]byte, 8*len(live))
	for i, d := range live {
		binary.LittleEndian.PutUint64(lb[i*8:], d)
	}
	return os.WriteFile(filepath.Join(dir, "live.bin"), lb, 0o644)
}

// Load 从目录加载段。
func Load(root string, id ID) (*Segment, error) {
	dir := filepath.Join(root, DirName(id))
	metaBytes, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, err
	}
	var meta Meta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, err
	}
	blob, err := os.ReadFile(filepath.Join(dir, "postings.bin"))
	if err != nil {
		return nil, err
	}
	postings := make(map[string]*posting.List)
	for len(blob) > 0 {
		if len(blob) < 4 {
			return nil, fmt.Errorf("segment: truncated term header")
		}
		termLen := binary.LittleEndian.Uint32(blob[:4])
		blob = blob[4:]
		if uint32(len(blob)) < termLen {
			return nil, fmt.Errorf("segment: truncated term")
		}
		term := string(blob[:termLen])
		blob = blob[termLen:]
		if len(blob) < 4 {
			return nil, fmt.Errorf("segment: truncated posting header")
		}
		encLen := binary.LittleEndian.Uint32(blob[:4])
		blob = blob[4:]
		if uint32(len(blob)) < encLen {
			return nil, fmt.Errorf("segment: truncated posting")
		}
		enc := blob[:encLen]
		blob = blob[encLen:]
		list, err := posting.Decode(enc)
		if err != nil {
			return nil, err
		}
		postings[term] = list
	}
	liveBytes, err := os.ReadFile(filepath.Join(dir, "live.bin"))
	if err != nil {
		return nil, err
	}
	live := make(map[uint64]struct{}, len(liveBytes)/8)
	for i := 0; i+8 <= len(liveBytes); i += 8 {
		live[binary.LittleEndian.Uint64(liveBytes[i:])] = struct{}{}
	}
	return &Segment{
		meta:      meta,
		postings:  postings,
		liveDocs:  live,
		tombstone: make(map[uint64]struct{}),
	}, nil
}
