package analyze

// DefaultEnglishStopwords 返回一组常见英文停用词。
func DefaultEnglishStopwords() map[string]struct{} {
	words := []string{
		"a", "an", "the", "and", "or", "but", "if", "then", "else", "when",
		"at", "by", "for", "with", "about", "against", "between", "into",
		"through", "during", "before", "after", "above", "below", "to", "from",
		"up", "down", "in", "out", "on", "off", "over", "under", "again",
		"further", "once", "here", "there", "all", "any", "both", "each",
		"few", "more", "most", "other", "some", "such", "no", "nor", "not",
		"only", "own", "same", "so", "than", "too", "very", "can", "will",
		"just", "don", "should", "now", "is", "are", "was", "were", "be",
		"been", "being", "have", "has", "had", "having", "do", "does", "did",
		"doing", "of", "as", "this", "that", "these", "those", "it", "its",
		"i", "you", "he", "she", "we", "they", "me", "him", "her", "us", "them",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}

// MergeStopwords 合并多组停用词。
func MergeStopwords(sets ...map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{})
	for _, s := range sets {
		for w := range s {
			out[w] = struct{}{}
		}
	}
	return out
}

// StopwordList 从切片构造集合。
func StopwordList(words ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		n := Normalize(w)
		if n == "" {
			continue
		}
		m[n] = struct{}{}
	}
	return m
}
