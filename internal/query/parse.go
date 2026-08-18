package query

import (
	"fmt"
	"strings"

	"github.com/LYH2263/go-logindex/internal/analyze"
)

// ParseSimple 解析极简查询语法：
//
//	term
//	term AND term
//	term OR term
//	NOT term
//	( ... )
//
// 仅支持大写 AND/OR/NOT 与括号，词项为标识符。
func ParseSimple(s string) (*Query, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("query: empty")
	}
	p := &parser{src: s}
	q, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.skipWS()
	if p.pos < len(p.src) {
		return nil, fmt.Errorf("query: trailing input at %d", p.pos)
	}
	return Optimize(q), nil
}

type parser struct {
	src string
	pos int
}

func (p *parser) skipWS() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.pos++
			continue
		}
		break
	}
}

func (p *parser) parseExpr() (*Query, error) {
	return p.parseOr()
}

func (p *parser) parseOr() (*Query, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWS()
		if !p.consumeKeyword("OR") {
			break
		}
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = Or(left, right)
	}
	return left, nil
}

func (p *parser) parseAnd() (*Query, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWS()
		if !p.consumeKeyword("AND") {
			break
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = And(left, right)
	}
	return left, nil
}

func (p *parser) parseUnary() (*Query, error) {
	p.skipWS()
	if p.consumeKeyword("NOT") {
		inner, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return Not(inner), nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (*Query, error) {
	p.skipWS()
	if p.pos >= len(p.src) {
		return nil, fmt.Errorf("query: unexpected end")
	}
	if p.src[p.pos] == '(' {
		p.pos++
		q, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.skipWS()
		if p.pos >= len(p.src) || p.src[p.pos] != ')' {
			return nil, fmt.Errorf("query: missing )")
		}
		p.pos++
		return q, nil
	}
	term, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	return Term(term), nil
}

func (p *parser) parseTerm() (string, error) {
	p.skipWS()
	start := p.pos
	for p.pos < len(p.src) {
		r := rune(p.src[p.pos])
		if analyze.IsIdentRune(r) {
			p.pos++
			continue
		}
		break
	}
	if p.pos == start {
		return "", fmt.Errorf("query: expected term at %d", p.pos)
	}
	return analyze.Normalize(p.src[start:p.pos]), nil
}

func (p *parser) consumeKeyword(kw string) bool {
	p.skipWS()
	if p.pos+len(kw) > len(p.src) {
		return false
	}
	if !strings.EqualFold(p.src[p.pos:p.pos+len(kw)], kw) {
		return false
	}
	end := p.pos + len(kw)
	if end < len(p.src) {
		r := rune(p.src[end])
		if analyze.IsIdentRune(r) {
			return false
		}
	}
	p.pos = end
	return true
}
