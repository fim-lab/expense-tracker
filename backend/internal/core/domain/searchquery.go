package domain

import (
	"fmt"
	"strings"
)

type searchQueryOp int

const (
	searchQueryTerm searchQueryOp = iota
	searchQueryAnd
	searchQueryOr
	searchQueryNot
)

type SearchQueryNode struct {
	op       searchQueryOp
	term     string
	children []*SearchQueryNode
}

type searchQueryTokenKind int

const (
	searchQueryTokenTerm searchQueryTokenKind = iota
	searchQueryTokenAnd
	searchQueryTokenOr
	searchQueryTokenNot
)

type searchQueryToken struct {
	kind searchQueryTokenKind
	term string
}

func tokenizeSearchQuery(query string) []searchQueryToken {
	var tokens []searchQueryToken
	var pending []string

	flush := func() {
		if len(pending) > 0 {
			tokens = append(tokens, searchQueryToken{kind: searchQueryTokenTerm, term: strings.Join(pending, " ")})
			pending = nil
		}
	}

	for _, word := range strings.Fields(query) {
		switch word {
		case "AND":
			flush()
			tokens = append(tokens, searchQueryToken{kind: searchQueryTokenAnd})
		case "OR":
			flush()
			tokens = append(tokens, searchQueryToken{kind: searchQueryTokenOr})
		case "NOT":
			flush()
			tokens = append(tokens, searchQueryToken{kind: searchQueryTokenNot})
		default:
			pending = append(pending, word)
		}
	}
	flush()

	return tokens
}

type searchQueryParser struct {
	tokens []searchQueryToken
	pos    int
}

func (p *searchQueryParser) peek() *searchQueryToken {
	if p.pos < len(p.tokens) {
		return &p.tokens[p.pos]
	}
	return nil
}

func (p *searchQueryParser) advance() *searchQueryToken {
	t := p.peek()
	if t != nil {
		p.pos++
	}
	return t
}

func (p *searchQueryParser) parseOr() *SearchQueryNode {
	node := p.parseAnd()
	for {
		t := p.peek()
		if t == nil || t.kind != searchQueryTokenOr {
			return node
		}
		p.advance()
		node = &SearchQueryNode{op: searchQueryOr, children: []*SearchQueryNode{node, p.parseAnd()}}
	}
}

func (p *searchQueryParser) parseAnd() *SearchQueryNode {
	node := p.parseNot()
	for {
		t := p.peek()
		if t == nil || t.kind != searchQueryTokenAnd {
			return node
		}
		p.advance()
		node = &SearchQueryNode{op: searchQueryAnd, children: []*SearchQueryNode{node, p.parseNot()}}
	}
}

func (p *searchQueryParser) parseNot() *SearchQueryNode {
	if t := p.peek(); t != nil && t.kind == searchQueryTokenNot {
		p.advance()
		return &SearchQueryNode{op: searchQueryNot, children: []*SearchQueryNode{p.parseNot()}}
	}
	return p.parsePrimary()
}

// parsePrimary consumes a term token. A dangling operator (a query ending
// in "AND", or two operators in a row) is treated as an empty term, which
// matches everything, rather than as a parse error.
func (p *searchQueryParser) parsePrimary() *SearchQueryNode {
	t := p.advance()
	if t == nil || t.kind != searchQueryTokenTerm {
		return &SearchQueryNode{op: searchQueryTerm}
	}
	return &SearchQueryNode{op: searchQueryTerm, term: t.term}
}

func ParseSearchQuery(query string) *SearchQueryNode {
	p := &searchQueryParser{tokens: tokenizeSearchQuery(query)}
	return p.parseOr()
}

func (n *SearchQueryNode) Matches(text string) bool {
	if n == nil {
		return true
	}
	switch n.op {
	case searchQueryNot:
		return !n.children[0].Matches(text)
	case searchQueryAnd:
		for _, c := range n.children {
			if !c.Matches(text) {
				return false
			}
		}
		return true
	case searchQueryOr:
		for _, c := range n.children {
			if c.Matches(text) {
				return true
			}
		}
		return false
	default:
		if n.term == "" {
			return true
		}
		return strings.Contains(strings.ToLower(text), strings.ToLower(n.term))
	}
}

func (n *SearchQueryNode) SQLCondition(column string, args *[]interface{}, argID *int) string {
	if n == nil {
		return "TRUE"
	}
	switch n.op {
	case searchQueryNot:
		return "NOT (" + n.children[0].SQLCondition(column, args, argID) + ")"
	case searchQueryAnd, searchQueryOr:
		joiner := " AND "
		if n.op == searchQueryOr {
			joiner = " OR "
		}
		parts := make([]string, len(n.children))
		for i, c := range n.children {
			parts[i] = c.SQLCondition(column, args, argID)
		}
		return "(" + strings.Join(parts, joiner) + ")"
	default:
		if n.term == "" {
			return "TRUE"
		}
		condition := fmt.Sprintf("%s ILIKE $%d", column, *argID)
		*args = append(*args, "%"+n.term+"%")
		*argID++
		return condition
	}
}
