package logindex

import "errors"

var (
	// ErrClosed 表示索引已关闭。
	ErrClosed = errors.New("logindex: index closed")
	// ErrEmptyQuery 表示查询为空。
	ErrEmptyQuery = errors.New("logindex: empty query")
	// ErrInvalidDocID 表示文档 ID 非法。
	ErrInvalidDocID = errors.New("logindex: invalid doc id")
	// ErrCorruptWAL 表示索引预写日志损坏。
	ErrCorruptWAL = errors.New("logindex: corrupt walix")
	// ErrSegmentNotFound 表示段不存在。
	ErrSegmentNotFound = errors.New("logindex: segment not found")
)
