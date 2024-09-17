package parser

import (
	"github.com/cyamas/rizz/internal/highlighter/lexer"
	"github.com/cyamas/rizz/internal/highlighter/token"
)

type Parser struct {
	lexer   *lexer.Lexer
	tokens  []token.Token
	context token.TokenType
}

func NewParser(l *lexer.Lexer) *Parser {
	return &Parser{lexer: l, tokens: []token.Token{}}
}

func (p *Parser) ParseLine(line string, context token.TokenType) {
	p.context = context
	p.clearTokens()
	p.lexer.LoadLine(line)

	for {
		nextToken := p.lexer.NextToken()
		if nextToken.Type == token.EOF {
			break
		}
		p.tokens = append(p.tokens, nextToken)
	}
	p.context = p.lexer.Context()
}

func (p *Parser) clearTokens() {
	p.tokens = []token.Token{}
}
