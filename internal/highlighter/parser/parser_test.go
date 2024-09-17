package parser

import (
	"fmt"
	"testing"

	"github.com/cyamas/rizz/internal/highlighter/lexer"
	"github.com/cyamas/rizz/internal/highlighter/token"
)

func TestParseLine(t *testing.T) {
	l := lexer.New()
	p := NewParser(l)

	input1 := "x := 5"
	expTokens1 := []token.Token{
		{Type: token.IDENT, Literal: "x"},
		{Type: token.SHORT_VAR_ASSIGN, Literal: ":="},
		{Type: token.INT_LITERAL, Literal: "5"},
	}

	input2 := "func add(x, y int) int {"
	expTokens2 := []token.Token{
		{Type: token.FUNC_DECLARE, Literal: "func"},
		{Type: token.FUNC_NAME, Literal: "add"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.PARAM_NAME, Literal: "x"},
		{Type: token.COMMA, Literal: ","},
		{Type: token.PARAM_NAME, Literal: "y"},
		{Type: token.INT, Literal: "int"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.RETURN_TYPE, Literal: "int"},
		{Type: token.LBRACE, Literal: "{"},
	}

	input3 := "return x + y"
	expTokens3 := []token.Token{
		{Type: token.RETURN, Literal: "return"},
		{Type: token.IDENT, Literal: "x"},
		{Type: token.PLUS, Literal: "+"},
		{Type: token.IDENT, Literal: "y"},
	}

	input4 := "}"
	expTokens4 := []token.Token{
		{Type: token.RBRACE, Literal: "}"},
	}

	input5 := "for i := range 10 {"
	expTokens5 := []token.Token{
		{Type: token.FOR, Literal: "for"},
		{Type: token.IDENT, Literal: "i"},
		{Type: token.SHORT_VAR_ASSIGN, Literal: ":="},
		{Type: token.RANGE, Literal: "range"},
		{Type: token.INT_LITERAL, Literal: "10"},
		{Type: token.LBRACE, Literal: "{"},
	}

	input6 := "if x == 5 {"
	expTokens6 := []token.Token{
		{Type: token.IF, Literal: "if"},
		{Type: token.IDENT, Literal: "x"},
		{Type: token.EQ, Literal: "=="},
		{Type: token.INT_LITERAL, Literal: "5"},
		{Type: token.LBRACE, Literal: "{"},
	}

	input7 := "return true"
	expTokens7 := []token.Token{
		{Type: token.RETURN, Literal: "return"},
		{Type: token.TRUE, Literal: "true"},
	}

	input8 := "} else {"
	expTokens8 := []token.Token{
		{Type: token.RBRACE, Literal: "}"},
		{Type: token.ELSE, Literal: "else"},
		{Type: token.LBRACE, Literal: "{"},
	}

	input9 := "return false"
	expTokens9 := []token.Token{
		{Type: token.RETURN, Literal: "return"},
		{Type: token.FALSE, Literal: "false"},
	}

	input10 := "}"
	expTokens10 := []token.Token{
		{Type: token.RBRACE, Literal: "}"},
	}

	input11 := "}"
	expTokens11 := []token.Token{
		{Type: token.RBRACE, Literal: "}"},
	}

	input12 := "var exMap = map[float32]bool {"
	expTokens12 := []token.Token{
		{Type: token.VAR_DECLARE, Literal: "var"},
		{Type: token.VAR_NAME, Literal: "exMap"},
		{Type: token.ASSIGN, Literal: "="},
		{Type: token.MAP_DECLARE, Literal: "map"},
		{Type: token.LBRACKET, Literal: "["},
		{Type: token.FLOAT_32, Literal: "float32"},
		{Type: token.RBRACKET, Literal: "]"},
		{Type: token.BOOL, Literal: "bool"},
		{Type: token.LBRACE, Literal: "{"},
	}

	input13 := `"a": true,`
	expTokens13 := []token.Token{
		{Type: token.DBL_QUOTE, Literal: "\""},
		{Type: token.STRING_LITERAL, Literal: "a"},
		{Type: token.DBL_QUOTE, Literal: "\""},
		{Type: token.COLON, Literal: ":"},
		{Type: token.TRUE, Literal: "true"},
		{Type: token.COMMA, Literal: ","},
	}

	input14 := `"b": false,`
	expTokens14 := []token.Token{
		{Type: token.DBL_QUOTE, Literal: "\""},
		{Type: token.STRING_LITERAL, Literal: "b"},
		{Type: token.DBL_QUOTE, Literal: "\""},
		{Type: token.COLON, Literal: ":"},
		{Type: token.FALSE, Literal: "false"},
		{Type: token.COMMA, Literal: ","},
	}

	input15 := "}"
	expTokens15 := []token.Token{
		{Type: token.RBRACE, Literal: "}"},
	}

	tests := []struct {
		input      string
		context    token.TokenType
		expTokens  []token.Token
		expContext token.TokenType
	}{
		{input1, token.TYPE_NONE, expTokens1, token.TYPE_NONE},
		{input2, token.TYPE_NONE, expTokens2, token.FUNC_BODY},
		{input3, token.FUNC_BODY, expTokens3, token.FUNC_BODY},
		{input4, token.FUNC_BODY, expTokens4, token.TYPE_NONE},
		{input5, token.TYPE_NONE, expTokens5, token.LOOP_BODY},
		{input6, token.LOOP_BODY, expTokens6, token.COND_BODY},
		{input7, token.COND_BODY, expTokens7, token.COND_BODY},
		{input8, token.COND_BODY, expTokens8, token.COND_BODY},
		{input9, token.COND_BODY, expTokens9, token.COND_BODY},
		{input10, token.COND_BODY, expTokens10, token.LOOP_BODY},
		{input11, token.LOOP_BODY, expTokens11, token.TYPE_NONE},
		{input12, token.TYPE_NONE, expTokens12, token.MAP_BODY},
		{input13, token.MAP_BODY, expTokens13, token.MAP_BODY},
		{input14, token.MAP_BODY, expTokens14, token.MAP_BODY},
		{input15, token.MAP_BODY, expTokens15, token.TYPE_NONE},
	}

	var currTok token.Token
	for i, tt := range tests {
		p.ParseLine(tt.input, tt.context)

		for j, tok := range p.tokens {
			currTok = tok
			if tt.expTokens[j].Type != tok.Type {
				fmt.Printf("FAIL TEST %d\n", i+1)
				fmt.Println("EXPECTED: ", tt.expTokens[j].Literal, "ACTUAL: ", tok.Literal)
				t.Fatalf("Expected %q. Got %q", tt.expTokens[j].Type, tok.Type)
			}
			if tt.expTokens[j].Literal != tok.Literal {
				fmt.Printf("FAIL TEST %d\n", i+1)
				t.Fatalf("Expected %s. Got %s", tt.expTokens[j].Literal, tok.Literal)

			}
		}
		if tt.expContext != p.context {
			fmt.Printf("FAIL TEST %d\n", i+1)
			fmt.Println("CURR TOK: ", currTok.Literal)
			fmt.Println("LEXER CONTEXT: ", l.Context())
			t.Fatalf("Context should be %q. Got %q", tt.expContext, p.context)
		}
	}
}
