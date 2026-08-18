package stats

// Snapshot 只读统计快照。
type Snapshot struct {
	NumDocs   int
	NumTerms  int
	AvgDocLen float64
	TopTerms  []TermCount
}

// Capture 从 TermFreq 与 DocStats 捕获快照。
func Capture(tf *TermFreq, ds *DocStats, topK int) Snapshot {
	s := Snapshot{}
	if tf != nil {
		s.NumDocs = tf.NumDocs()
		s.TopTerms = tf.TopTerms(topK)
		s.NumTerms = len(s.TopTerms)
		// 更准确的词项数：
		tf.mu.RLock()
		s.NumTerms = len(tf.tf)
		tf.mu.RUnlock()
	}
	if ds != nil {
		s.AvgDocLen = ds.AvgLength()
		if s.NumDocs == 0 {
			s.NumDocs = ds.Count()
		}
	}
	return s
}

// IDF 简易逆文档频率：log((N+1)/(df+1)) + 1。
func IDF(numDocs, df int) float64 {
	if df < 0 {
		df = 0
	}
	if numDocs < 0 {
		numDocs = 0
	}
	return ln(float64(numDocs+1)/float64(df+1)) + 1
}

func ln(x float64) float64 {
	// 简单自然对数近似（足够用于排序）。
	if x <= 0 {
		return 0
	}
	// 使用 math 会更精确，这里保持无额外依赖也可；改用标准库。
	return logNatural(x)
}
