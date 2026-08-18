package logindex

import (
	"github.com/LYH2263/go-logindex/internal/analyze"
	"github.com/LYH2263/go-logindex/internal/merge"
)

// Options 控制索引行为。
type Options struct {
	Analyzer       analyze.Analyzer
	FlushThreshold  int
	MergePolicy     merge.Policy
	EnableWAL       bool
	EnableStats     bool
	MaxSearchHits   int
}

// Option 是功能性配置项。
type Option func(*Options)

// DefaultOptions 返回合理默认值。
func DefaultOptions() Options {
	return Options{
		Analyzer:       analyze.NewDefault(),
		FlushThreshold: 256,
		MergePolicy:     merge.DefaultPolicy(),
		EnableWAL:       true,
		EnableStats:     true,
		MaxSearchHits:   10_000,
	}
}

// WithAnalyzer 设置分词器。
func WithAnalyzer(a analyze.Analyzer) Option {
	return func(o *Options) { o.Analyzer = a }
}

// WithFlushThreshold 设置内存索引刷盘阈值（文档数）。
func WithFlushThreshold(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.FlushThreshold = n
		}
	}
}

// WithMergePolicy 设置段合并策略。
func WithMergePolicy(p merge.Policy) Option {
	return func(o *Options) { o.MergePolicy = p }
}

// WithWAL 开关预写日志。
func WithWAL(enabled bool) Option {
	return func(o *Options) { o.EnableWAL = enabled }
}

// WithStats 开关词频统计。
func WithStats(enabled bool) Option {
	return func(o *Options) { o.EnableStats = enabled }
}

// WithMaxSearchHits 限制单次检索返回条数。
func WithMaxSearchHits(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.MaxSearchHits = n
		}
	}
}

func applyOptions(opts []Option) Options {
	cfg := DefaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.Analyzer == nil {
		cfg.Analyzer = analyze.NewDefault()
	}
	if cfg.FlushThreshold <= 0 {
		cfg.FlushThreshold = 256
	}
	if cfg.MaxSearchHits <= 0 {
		cfg.MaxSearchHits = 10_000
	}
	return cfg
}
