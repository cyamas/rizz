package lexer

import (
	"fmt"
	"testing"

	"github.com/cyamas/rizz/internal/highlighter/token"
)

func TestNextToken(t *testing.T) {
	input := `var five int = 5`

	tests := []struct {
		expType     token.TokenType
		expLit      string
		expStartIdx int
		expLength   int
	}{
		{token.VAR_DECLARE, "var", 0, 3},
		{token.VAR_NAME, "five", 4, 4},
		{token.INT, "int", 9, 3},
		{token.ASSIGN, "=", 13, 1},
		{token.INT_LITERAL, "5", 15, 1},
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

		if tok.StartIndex != tt.expStartIdx {
			t.Fatalf("tok start index should be : %d. Got %d", tt.expStartIdx, tok.StartIndex)
		}
		if tok.Length != tt.expLength {
			t.Fatalf("TEST %d: tok length should be : %d. Got %d", i, tt.expLength, tok.Length)
		}
	}

}
