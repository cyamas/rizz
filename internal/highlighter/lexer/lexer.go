package lexer

import (
	"github.com/cyamas/rizz/internal/highlighter/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	context      []token.TokenType
}

func New() *Lexer {
	l := &Lexer{}
	l.context = []token.TokenType{token.TYPE_NONE}
	return l
}

func (l *Lexer) LoadLine(input string) {
	l.input = input
	l.position = 0
	l.readPosition = 0
	l.readChar()
}

func (l *Lexer) Context() token.TokenType {
	return l.context[len(l.context)-1]
}

func (l *Lexer) addContext(tokType token.TokenType) {
	l.context = append(l.context, tokType)
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
			tok.SetIndex(l.position - 1)
			tok.SetLength(2)
		} else {
			tok = newToken(token.ASSIGN, l.ch, l.position, 1)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, l.position, 1)
	case '(':
		tok = newToken(token.LPAREN, l.ch, l.position, 1)
	case ')':
		if l.Context() == token.START_PARAMS {
			l.ReplaceContext(token.END_PARAMS)
		}
		tok = newToken(token.RPAREN, l.ch, l.position, 1)
	case ',':
		tok = newToken(token.COMMA, l.ch, l.position, 1)
	case '+':
		tok = newToken(token.PLUS, l.ch, l.position, 1)
	case '-':
		tok = newToken(token.MINUS, l.ch, l.position, 1)
	case '*':
		tok = newToken(token.ASTERISK, l.ch, l.position, 1)
	case '/':
		tok = newToken(token.SLASH, l.ch, l.position, 1)
	case '<':
		tok = newToken(token.LT, l.ch, l.position, 1)
	case '>':
		tok = newToken(token.GT, l.ch, l.position, 1)
	case '{':
		if l.Context() == token.FUNC_SIG {
			l.ReplaceContext(token.FUNC_BODY)
		}
		if l.Context() == token.FOR_SIG {
			l.ReplaceContext(token.LOOP_BODY)
		}
		if l.Context() == token.IF_SIG {
			l.ReplaceContext(token.COND_BODY)
		}
		if l.Context() == token.ELSE {
			l.ReplaceContext(token.COND_BODY)
		}
		if l.Context() == token.MAP_SIG {
			l.ReplaceContext(token.MAP_BODY)
		}
		tok = newToken(token.LBRACE, l.ch, l.position, 1)
	case '}':
		if l.Context() == token.MAP_BODY {
			l.RemoveContext()
		}
		if l.Context() == token.FUNC_BODY {
			l.RemoveContext()
		}
		if l.Context() == token.LOOP_BODY {
			l.RemoveContext()
		}
		if l.Context() == token.COND_BODY {
			l.RemoveContext()
		}
		tok = newToken(token.RBRACE, l.ch, l.position, 1)
	case '[':
		tok = newToken(token.LBRACKET, l.ch, l.position, 1)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, l.position, 1)
	case '"':
		if l.Context() == token.DBL_QUOTE {
			l.RemoveContext()
		} else {
			l.addContext(token.DBL_QUOTE)
		}
		tok = newToken(token.DBL_QUOTE, l.ch, l.position, 1)
	case ':':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.SHORT_VAR_ASSIGN, Literal: literal}
			tok.SetIndex(l.position - 1)
			tok.SetLength(2)

		} else {
			tok = newToken(token.COLON, l.ch, l.position, 1)
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		switch {
		case isLetter(l.ch):
			start := l.position
			tok.SetIndex(start)
			tok.Literal = l.readIdentifier()
			tok.SetLength(l.position - start)
			tok.Type = token.LookupIdent(tok.Literal, l.Context())
			if tok.Type == token.MAP_DECLARE {
				l.addContext(token.MAP_SIG)
			}
			if tok.Type == token.IF {
				l.addContext(token.IF_SIG)
			}
			if tok.Type == token.ELSE {
				l.addContext(token.ELSE)
			}
			if tok.Type == token.FOR {
				l.addContext(token.FOR_SIG)
			}
			if tok.Type == token.RETURN_TYPE {
				l.RemoveContext()
				l.ReplaceContext(token.FUNC_BODY)
			}
			if tok.Type == token.FUNC_NAME {
				l.ReplaceContext(token.START_PARAMS)
			}
			if tok.Type == token.STRUCT {
				l.addContext(token.STRUCT)
			}
			if tok.Type == token.VAR_DECLARE {
				l.addContext(token.VAR_DECLARE)
			}
			if tok.Type == token.VAR_NAME {
				l.RemoveContext()
			}
			if tok.Type == token.FUNC_DECLARE {
				l.addContext(token.FUNC_SIG)
				l.addContext(token.FUNC_DECLARE)
			}
			return tok
		case isDigit(l.ch):
			tok.Type = token.INT_LITERAL
			tok.Literal = l.readNumber()
			return tok
		default:
			tok = newToken(token.ILLEGAL, 0, l.position, 1)
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) RemoveContext() {
	l.context = l.context[:len(l.context)-1]
}

func (l *Lexer) ReplaceContext(context token.TokenType) {
	l.context[len(l.context)-1] = context
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func newToken(tokenType token.TokenType, ch byte, idx int, length int) token.Token {
	tok := token.Token{Type: tokenType, Literal: string(ch)}
	tok.SetIndex(idx)
	tok.SetLength(length)
	return tok
}
