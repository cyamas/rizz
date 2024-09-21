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

func (l *Lexer) SetContext(context []token.TokenType) {
	l.context = context
}

func (l *Lexer) Context() token.TokenType {
	return l.context[len(l.context)-1]
}

func (l *Lexer) AddContext(tokType token.TokenType) {
	l.context = append(l.context, tokType)
}

func (l *Lexer) LineContext() []token.TokenType {
	return l.context
}

func (l *Lexer) Clear() {
	l.input = ""
	l.position = 0
	l.readPosition = 0
	l.ch = 0
	l.context = nil
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
			l.AddContext(token.ASSIGN)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, l.position, 1)
	case '(':
		tok = newToken(token.LPAREN, l.ch, l.position, 1)
	case ')':
		if l.Context() == token.START_PARAMS {
			l.ReplaceContext(token.END_PARAMS)
		}
		if l.Context() == token.MULTI_IMPORT {
			l.RemoveContext()
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
		switch l.Context() {
		case token.SLICE_DECLARE:
			l.ReplaceContext(token.SLICE_BODY)
		case token.FUNC_SIG:
			l.ReplaceContext(token.FUNC_BODY)
		case token.FOR_SIG:
			l.ReplaceContext(token.LOOP_BODY)
		case token.MAP_SIG:
			l.ReplaceContext(token.MAP_BODY)
		case token.IF_SIG:
			l.ReplaceContext(token.COND_BODY)
		case token.ELSE:
			l.ReplaceContext(token.COND_BODY)
		case token.VAL_DECLARE:
			l.ReplaceContext(token.MAP_BODY)
		}
		tok = newToken(token.LBRACE, l.ch, l.position, 1)
	case '}':
		if l.contextIsBody() {
			l.RemoveContext()
		}
		tok = newToken(token.RBRACE, l.ch, l.position, 1)
	case '[':
		if l.Context() == token.ASSIGN {
			l.RemoveContext()
			literal := string(l.ch)
			peekChar := l.peekChar()
			if peekChar == ']' {
				literal += "]"
				tok = token.Token{Type: token.SLICE_DECLARE, Literal: literal}
				tok.SetIndex(l.position)
				tok.SetLength(2)
				l.AddContext(token.SLICE_DECLARE)
				l.readChar()
			}
			if isDigit(peekChar) {
				start := l.position
				l.readChar()
				literal += l.readNumber()
				if l.ch == ']' {
					literal += "]"
				}
				tok = token.Token{Type: token.ARRAY_DECLARE, Literal: literal}
				tok.SetIndex(start)
				tok.SetLength(l.readPosition - start)
				l.AddContext(token.ARRAY_DECLARE)
			}
		} else {
			tok = newToken(token.LBRACKET, l.ch, l.position, 1)
		}
	case ']':
		tok = newToken(token.RBRACKET, l.ch, l.position, 1)
	case '\'':
		if l.Context() == token.ASSIGN {
			l.RemoveContext()
		}
		switch l.Context() {
		case token.RUNE:
			l.ReplaceContext(token.RUNE_START)
		case token.RUNE_START:
			l.RemoveContext()
		}
		tok = newToken(token.SINGLE_QUOTE, l.ch, l.position, 1)
	case '"':
		switch l.Context() {
		case token.ASSIGN:
			l.RemoveContext()
		case token.MAP_BODY:
			tok = l.createStringKeyToken()
			return tok
		case token.SINGLE_IMPORT:
			l.ReplaceContext(token.START_SINGLE_IMPORT)
			l.AddContext(token.START_IMPORT_NAME)
		case token.START_SINGLE_IMPORT:
			l.RemoveContext()
		case token.MULTI_IMPORT:
			l.AddContext(token.START_IMPORT_NAME)
		case token.START_IMPORT_NAME:
			l.RemoveContext()
		case token.DBL_QUOTE:
			l.RemoveContext()
		default:
			l.AddContext(token.DBL_QUOTE)
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
			l.AddContext(token.ASSIGN)
		} else {
			tok = newToken(token.COLON, l.ch, l.position, 1)
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		switch {
		case isLetter(l.ch):
			if l.Context() == token.ASSIGN {
				l.RemoveContext()
			}
			start := l.position
			tok.SetIndex(start)
			tok.Literal = l.readIdentifier()
			tok.SetLength(l.position - start)
			tok.Type = token.LookupIdent(tok.Literal, l.Context())
			switch tok.Type {
			case token.TYPE:
				l.AddContext(token.TYPE)
			case token.PACKAGE:
				l.AddContext(token.PACKAGE)
			case token.PACKAGE_NAME:
				l.RemoveContext()
			case token.IMPORT:
				if l.peekChar() == '(' {
					l.AddContext(token.MULTI_IMPORT)
				} else {
					l.AddContext(token.SINGLE_IMPORT)
				}
			case token.IMPORT_NAME:
				l.RemoveContext()
			case token.ARRAY_TYPE:
				if l.peekChar() == 0 {
					l.RemoveContext()
				}
			case token.VALUE:
				l.RemoveContext()
			case token.VAL_DECLARE:
				l.ReplaceContext(token.VAL_DECLARE)
			case token.KEY_DECLARE:
				l.ReplaceContext(token.KEY_DECLARE)
			case token.RUNE:
				l.AddContext(token.RUNE)
			case token.MAP_DECLARE:
				l.AddContext(token.MAP_SIG)
			case token.IF:
				l.AddContext(token.IF_SIG)
			case token.ELSE:
				l.AddContext(token.ELSE)
			case token.FOR:
				l.AddContext(token.FOR_SIG)
			case token.RETURN_TYPE:
				l.RemoveContext()
				l.ReplaceContext(token.FUNC_BODY)
			case token.FUNC_NAME:
				l.ReplaceContext(token.START_PARAMS)
			case token.PARAM_NAME:
				if isLetter(l.peekChar()) {
					l.AddContext(token.PARAM_NAME)
				}
			case token.PARAM_TYPE:
				l.RemoveContext()
			case token.STRUCT:
				l.AddContext(token.STRUCT)
			case token.VAR_DECLARE:
				l.AddContext(token.VAR_DECLARE)
			case token.VAR_NAME:

				l.ReplaceContext(token.ASSIGN)
			case token.FUNC_DECLARE:
				l.AddContext(token.FUNC_SIG)
				l.AddContext(token.FUNC_DECLARE)
			}
			return tok
		case isDigit(l.ch):
			if l.Context() == token.ASSIGN {
				l.RemoveContext()
			}
			start := l.position
			tok.SetIndex(start)
			tok.Type = token.INT_LITERAL
			literal := l.readNumber()
			if l.ch == '.' {
				l.readChar()
				tok.Type = token.FLOAT_LITERAL
				literal += "."
				literal += l.readNumber()
			}
			tok.Literal = literal
			tok.SetLength(l.position - start)
			switch l.Context() {
			case token.DBL_QUOTE:
				tok.Type = token.STRING_LITERAL
			case token.SINGLE_QUOTE:
				tok.Type = token.RUNE_LITERAL
			case token.SLICE_BODY:
				tok.Type = token.ITEM
			case token.MAP_BODY:
				tok.Type = token.KEY
			case token.KEY:
				tok.Type = token.VALUE
			}
			return tok
		default:
			tok = newToken(token.ILLEGAL, 0, l.position, 1)
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) contextIsBody() bool {
	ctx := l.Context()
	return ctx == token.FUNC_BODY || ctx == token.COND_BODY || ctx == token.LOOP_BODY || ctx == token.MAP_BODY || ctx == token.SLICE_BODY
}

func (l *Lexer) createStringKeyToken() token.Token {
	start := l.position
	literal := string(l.ch)
	l.readChar()
	literal += l.readIdentifier()
	literal += string(l.ch)
	tok := token.Token{Type: token.KEY, Literal: literal}
	l.readChar()
	length := l.position - start
	tok.SetIndex(start)
	tok.SetLength(length)
	l.AddContext(token.KEY)
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
		charPos := l.readPosition
		for l.input[charPos] == ' ' || l.input[charPos] == '\t' {
			charPos++
		}
		return l.input[charPos]
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
