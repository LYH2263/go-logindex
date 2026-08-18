package logindex

import "github.com/LYH2263/go-logindex/internal/query"

// DocID 是文档唯一标识。
type DocID = uint64

// Hit 是一次检索命中。
type Hit struct {
	DocID DocID
	Score float64
}

// Query 是布尔查询表达式指针的别名。
type Query = *query.Query

// Term 构造词项查询。
func Term(term string) Query { return query.Term(term) }

// And 构造合取查询。
func And(parts ...Query) Query { return query.And(parts...) }

// Or 构造析取查询。
func Or(parts ...Query) Query { return query.Or(parts...) }

// Not 构造否定查询（相对全集 / 候选集）。
func Not(inner Query) Query { return query.Not(inner) }
