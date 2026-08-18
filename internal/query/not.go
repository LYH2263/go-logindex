package query

import "github.com/LYH2263/go-logindex/internal/posting"

// NOT 语义：Universe \ inner。
// 若无全集，则返回空（调用方应提供 Universe）。
func evalNot(p Planner, kids []*Query) (*posting.List, error) {
	if len(kids) == 0 || kids[0] == nil {
		return posting.New(), nil
	}
	inner, err := p.eval(kids[0])
	if err != nil {
		return nil, err
	}
	univ := p.Src.Universe()
	if univ == nil || univ.Len() == 0 {
		return posting.New(), nil
	}
	return posting.Difference(univ, inner), nil
}
