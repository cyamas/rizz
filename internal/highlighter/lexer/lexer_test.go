package lexer

import (
	"fmt"
	"testing"

	"github.com/cyamas/rizz/internal/highlighter/token"
)

func TestNextToken(t *testing.T) {
	input := `var five int = 5`

	tests := []struct {
		expType token.TokenType
		expLit  string
	}{
		{token.VAR_DECLARE, "var"},
		{token.VAR_NAME, "five"},
		{token.INT, "int"},
		{token.ASSIGN, "="},
		{token.INT_LITERAL, "5"},
	}

	l := New()
	l.LoadLine(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expType {
			fmt.Println("TEST ", i, "exp:", tt.expLit, "actual: ", tok.Literal, "context: ", l.Context())
			t.Fatalf("TEST %d: incorrect tokentype. Expected %q. Got %q", i, tt.expType, tok.Type)
		}
		if tok.Literal != tt.expLit {
			t.Fatalf("TEST %d: incorrect literal. Expected %s. Got %s", i, tt.expLit, tok.Literal)
		}
	}
}
