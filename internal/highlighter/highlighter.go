package highlighter

import (
	"github.com/cyamas/rizz/internal/highlighter/lexer"
	"github.com/cyamas/rizz/internal/highlighter/token"
)

type Highlighter struct {
	lexer       *lexer.Lexer
	tokens      []token.Token
	context     token.TokenType
	varNames    map[string]bool
	funcNames   map[string]bool
	importNames map[string]bool
	typeNames   map[string]bool
}

func New(l *lexer.Lexer) *Highlighter {
	return &Highlighter{
		lexer:       l,
		tokens:      []token.Token{},
		varNames:    make(map[string]bool),
		funcNames:   make(map[string]bool),
		importNames: make(map[string]bool),
		typeNames:   make(map[string]bool),
	}
}

func (h *Highlighter) LineContext() []token.TokenType {
	return h.lexer.LineContext()
}

func (h *Highlighter) Tokens() []token.Token {
	return h.tokens
}

func (h *Highlighter) ParseLine(line string, context []token.TokenType) {
	h.lexer.Clear()
	h.lexer.SetContext(context)
	h.clearTokens()
	h.lexer.LoadLine(line)

	for {
		nextToken := h.lexer.NextToken()
		if nextToken.Type == token.EOF {
			break
		}
		switch nextToken.Type {
		case token.IMPORT_NAME:
			h.importNames[nextToken.Literal] = true
		case token.TYPE_NAME:
			h.typeNames[nextToken.Literal] = true
		case token.VAR_NAME:
			h.varNames[nextToken.Literal] = true
		case token.FUNC_NAME:
			h.funcNames[nextToken.Literal] = true
		case token.IDENT:
			if _, ok := h.varNames[nextToken.Literal]; ok {
				nextToken.Type = token.VAR_CALL
			}
			if _, ok := h.funcNames[nextToken.Literal]; ok {
				nextToken.Type = token.FUNC_CALL
			}
			if _, ok := h.importNames[nextToken.Literal]; ok {
				nextToken.Type = token.IMPORT_CALL
			}
			if _, ok := h.typeNames[nextToken.Literal]; ok {
				nextToken.Type = token.TYPE_CALL
			}
		}
		h.tokens = append(h.tokens, nextToken)
	}
	h.context = h.lexer.Context()
}

func (h *Highlighter) clearTokens() {
	h.tokens = []token.Token{}
}
