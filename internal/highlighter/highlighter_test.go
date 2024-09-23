package highlighter

import (
	"fmt"
	"testing"

	"github.com/cyamas/rizz/internal/highlighter/lexer"
	"github.com/cyamas/rizz/internal/highlighter/token"
)

func TestParseLine(t *testing.T) {
	l := lexer.New()
	h := New(l)

	input1 := "x := 5"
	expTokens1 := []token.Token{
		{Type: token.IDENT, Literal: "x", StartIndex: 0, Length: 1},
		{Type: token.SHORT_VAR_ASSIGN, Literal: ":=", StartIndex: 2, Length: 2},
		{Type: token.INT_LITERAL, Literal: "5", StartIndex: 5, Length: 1},
	}
	ctx1 := []token.TokenType{token.TYPE_NONE}

	input2 := "func add(x, y int) int {"
	expTokens2 := []token.Token{
		{Type: token.FUNC_DECLARE, Literal: "func", StartIndex: 0, Length: 4},
		{Type: token.FUNC_NAME, Literal: "add", StartIndex: 5, Length: 3},
		{Type: token.LPAREN, Literal: "(", StartIndex: 8, Length: 1},
		{Type: token.PARAM_NAME, Literal: "x", StartIndex: 9, Length: 1},
		{Type: token.COMMA, Literal: ",", StartIndex: 10, Length: 1},
		{Type: token.PARAM_NAME, Literal: "y", StartIndex: 12, Length: 1},
		{Type: token.PARAM_TYPE, Literal: "int", StartIndex: 14, Length: 3},
		{Type: token.RPAREN, Literal: ")", StartIndex: 17, Length: 1},
		{Type: token.RETURN_TYPE, Literal: "int", StartIndex: 19, Length: 3},
		{Type: token.LBRACE, Literal: "{", StartIndex: 23, Length: 1},
	}
	ctx2 := []token.TokenType{token.TYPE_NONE, token.FUNC_BODY}

	input3 := "return x + y"
	expTokens3 := []token.Token{
		{Type: token.RETURN, Literal: "return", StartIndex: 0, Length: 6},
		{Type: token.IDENT, Literal: "x", StartIndex: 7, Length: 1},
		{Type: token.PLUS, Literal: "+", StartIndex: 9, Length: 1},
		{Type: token.IDENT, Literal: "y", StartIndex: 11, Length: 1},
	}
	ctx3 := []token.TokenType{token.TYPE_NONE, token.FUNC_BODY}

	input4 := "}"
	expTokens4 := []token.Token{
		{Type: token.RBRACE, Literal: "}", StartIndex: 0, Length: 1},
	}
	ctx4 := []token.TokenType{token.TYPE_NONE}

	input5 := "for i := range 10 {"
	expTokens5 := []token.Token{
		{Type: token.FOR, Literal: "for", StartIndex: 0, Length: 3},
		{Type: token.IDENT, Literal: "i", StartIndex: 4, Length: 1},
		{Type: token.SHORT_VAR_ASSIGN, Literal: ":=", StartIndex: 6, Length: 2},
		{Type: token.RANGE, Literal: "range", StartIndex: 9, Length: 5},
		{Type: token.INT_LITERAL, Literal: "10", StartIndex: 15, Length: 2},
		{Type: token.LBRACE, Literal: "{", StartIndex: 18, Length: 1},
	}
	ctx5 := []token.TokenType{token.TYPE_NONE, token.LOOP_BODY}

	input6 := "if x == 5 {"
	expTokens6 := []token.Token{
		{Type: token.IF, Literal: "if", StartIndex: 0, Length: 2},
		{Type: token.IDENT, Literal: "x", StartIndex: 3, Length: 1},
		{Type: token.EQ, Literal: "==", StartIndex: 5, Length: 2},
		{Type: token.INT_LITERAL, Literal: "5", StartIndex: 8, Length: 1},
		{Type: token.LBRACE, Literal: "{", StartIndex: 10, Length: 1},
	}
	ctx6 := []token.TokenType{token.TYPE_NONE, token.LOOP_BODY, token.COND_BODY}

	input7 := "return true"
	expTokens7 := []token.Token{
		{Type: token.RETURN, Literal: "return", StartIndex: 0, Length: 6},
		{Type: token.TRUE, Literal: "true", StartIndex: 7, Length: 4},
	}
	ctx7 := []token.TokenType{token.TYPE_NONE, token.COND_BODY}

	input8 := "} else {"
	expTokens8 := []token.Token{
		{Type: token.RBRACE, Literal: "}", StartIndex: 0, Length: 1},
		{Type: token.ELSE, Literal: "else", StartIndex: 2, Length: 4},
		{Type: token.LBRACE, Literal: "{", StartIndex: 7, Length: 1},
	}
	ctx8 := []token.TokenType{token.TYPE_NONE, token.LOOP_BODY, token.COND_BODY}

	input9 := "return false"
	expTokens9 := []token.Token{
		{Type: token.RETURN, Literal: "return", StartIndex: 0, Length: 6},
		{Type: token.FALSE, Literal: "false", StartIndex: 7, Length: 5},
	}
	ctx9 := []token.TokenType{token.TYPE_NONE, token.LOOP_BODY, token.COND_BODY}

	input10 := "}"
	expTokens10 := []token.Token{
		{Type: token.RBRACE, Literal: "}", StartIndex: 0, Length: 1},
	}
	ctx10 := []token.TokenType{token.TYPE_NONE, token.LOOP_BODY, token.COND_BODY}

	input11 := "}"
	expTokens11 := []token.Token{
		{Type: token.RBRACE, Literal: "}", StartIndex: 0, Length: 1},
	}
	ctx11 := []token.TokenType{token.TYPE_NONE, token.LOOP_BODY}

	input12 := "var exMap = map[float32]bool {"
	expTokens12 := []token.Token{
		{Type: token.VAR_DECLARE, Literal: "var", StartIndex: 0, Length: 3},
		{Type: token.VAR_NAME, Literal: "exMap", StartIndex: 4, Length: 5},
		{Type: token.ASSIGN, Literal: "=", StartIndex: 10, Length: 1},
		{Type: token.MAP_DECLARE, Literal: "map", StartIndex: 12, Length: 3},
		{Type: token.LBRACKET, Literal: "[", StartIndex: 15, Length: 1},
		{Type: token.KEY_DECLARE, Literal: "float32", StartIndex: 16, Length: 7},
		{Type: token.RBRACKET, Literal: "]", StartIndex: 23, Length: 1},
		{Type: token.VAL_DECLARE, Literal: "bool", StartIndex: 24, Length: 4},
		{Type: token.LBRACE, Literal: "{", StartIndex: 29, Length: 1},
	}
	ctx12 := []token.TokenType{token.TYPE_NONE, token.MAP_BODY}

	input13 := `"a": true,`
	expTokens13 := []token.Token{
		{Type: token.KEY, Literal: `"a"`, StartIndex: 0, Length: 3},
		{Type: token.COLON, Literal: ":", StartIndex: 3, Length: 1},
		{Type: token.VALUE, Literal: "true", StartIndex: 5, Length: 4},
		{Type: token.COMMA, Literal: ",", StartIndex: 9, Length: 1},
	}
	ctx13 := []token.TokenType{token.TYPE_NONE, token.MAP_BODY}

	input14 := `"b": false,`
	expTokens14 := []token.Token{
		{Type: token.KEY, Literal: `"b"`, StartIndex: 0, Length: 3},
		{Type: token.COLON, Literal: ":", StartIndex: 3, Length: 1},
		{Type: token.VALUE, Literal: "false", StartIndex: 5, Length: 5},
		{Type: token.COMMA, Literal: ",", StartIndex: 10, Length: 1},
	}
	ctx14 := []token.TokenType{token.TYPE_NONE, token.MAP_BODY}

	input15 := "}"
	expTokens15 := []token.Token{
		{Type: token.RBRACE, Literal: "}", StartIndex: 0, Length: 1},
	}
	ctx15 := []token.TokenType{token.TYPE_NONE}

	input16 := "var fl64 float64"
	expTokens16 := []token.Token{
		{Type: token.VAR_DECLARE, Literal: "var", StartIndex: 0, Length: 3},
		{Type: token.VAR_NAME, Literal: "fl64", StartIndex: 4, Length: 4},
		{Type: token.FLOAT_64, Literal: "float64", StartIndex: 9, Length: 7},
	}
	ctx16 := []token.TokenType{token.TYPE_NONE}

	input17 := "fl64 = 1.002"
	expTokens17 := []token.Token{
		{Type: token.VAR_CALL, Literal: "fl64", StartIndex: 0, Length: 4},
		{Type: token.ASSIGN, Literal: "=", StartIndex: 5, Length: 1},
		{Type: token.FLOAT_LITERAL, Literal: "1.002", StartIndex: 7, Length: 5},
	}
	ctx17 := []token.TokenType{token.TYPE_NONE}

	input18 := "var z rune = 'z'"
	expTokens18 := []token.Token{
		{Type: token.VAR_DECLARE, Literal: "var", StartIndex: 0, Length: 3},
		{Type: token.VAR_NAME, Literal: "z", StartIndex: 4, Length: 1},
		{Type: token.RUNE, Literal: "rune", StartIndex: 6, Length: 4},
		{Type: token.ASSIGN, Literal: "=", StartIndex: 11, Length: 1},
		{Type: token.SINGLE_QUOTE, Literal: "'", StartIndex: 13, Length: 1},
		{Type: token.RUNE_LITERAL, Literal: "z", StartIndex: 14, Length: 1},
		{Type: token.SINGLE_QUOTE, Literal: "'", StartIndex: 15, Length: 1},
	}
	ctx18 := []token.TokenType{token.TYPE_NONE}

	input19 := "make interface any uint8 package main import \"fmt\""
	expTokens19 := []token.Token{
		{Type: token.MAKE, Literal: "make", StartIndex: 0, Length: 4},
		{Type: token.INTERFACE, Literal: "interface", StartIndex: 5, Length: 9},
		{Type: token.ANY, Literal: "any", StartIndex: 15, Length: 3},
		{Type: token.UINT_8, Literal: "uint8", StartIndex: 19, Length: 5},
		{Type: token.PACKAGE, Literal: "package", StartIndex: 25, Length: 7},
		{Type: token.PACKAGE_NAME, Literal: "main", StartIndex: 33, Length: 4},
		{Type: token.IMPORT, Literal: "import", StartIndex: 38, Length: 6},
		{Type: token.DBL_QUOTE, Literal: "\"", StartIndex: 45, Length: 1},
		{Type: token.IMPORT_NAME, Literal: "fmt", StartIndex: 46, Length: 3},
		{Type: token.DBL_QUOTE, Literal: "\"", StartIndex: 49, Length: 1},
	}
	ctx19 := []token.TokenType{token.TYPE_NONE}

	input20 := "byteMap := map[byte]interface {}"
	expTokens20 := []token.Token{
		{Type: token.IDENT, Literal: "byteMap", StartIndex: 0, Length: 7},
		{Type: token.SHORT_VAR_ASSIGN, Literal: ":=", StartIndex: 8, Length: 2},
		{Type: token.MAP_DECLARE, Literal: "map", StartIndex: 11, Length: 3},
		{Type: token.LBRACKET, Literal: "[", StartIndex: 14, Length: 1},
		{Type: token.KEY_DECLARE, Literal: "byte", StartIndex: 15, Length: 4},
		{Type: token.RBRACKET, Literal: "]", StartIndex: 19, Length: 1},
		{Type: token.VAL_DECLARE, Literal: "interface", StartIndex: 20, Length: 9},
		{Type: token.LBRACE, Literal: "{", StartIndex: 30, Length: 1},
		{Type: token.RBRACE, Literal: "}", StartIndex: 31, Length: 1},
	}
	ctx20 := []token.TokenType{token.TYPE_NONE}

	input21 := "exSlice := []uint8 {0, 1}"
	expTokens21 := []token.Token{
		{Type: token.IDENT, Literal: "exSlice", StartIndex: 0, Length: 7},
		{Type: token.SHORT_VAR_ASSIGN, Literal: ":=", StartIndex: 8, Length: 2},
		{Type: token.SLICE_DECLARE, Literal: "[]", StartIndex: 11, Length: 2},
		{Type: token.ITEM_TYPE, Literal: "uint8", StartIndex: 13, Length: 5},
		{Type: token.LBRACE, Literal: "{", StartIndex: 19, Length: 1},
		{Type: token.ITEM, Literal: "0", StartIndex: 20, Length: 1},
		{Type: token.COMMA, Literal: ",", StartIndex: 21, Length: 1},
		{Type: token.ITEM, Literal: "1", StartIndex: 23, Length: 1},
		{Type: token.RBRACE, Literal: "}", StartIndex: 24, Length: 1},
	}
	ctx21 := []token.TokenType{token.TYPE_NONE}

	input22 := "var arr [3]string"
	expTokens22 := []token.Token{
		{Type: token.VAR_DECLARE, Literal: "var", StartIndex: 0, Length: 3},
		{Type: token.VAR_NAME, Literal: "arr", StartIndex: 4, Length: 3},
		{Type: token.ARRAY_DECLARE, Literal: "[3]", StartIndex: 8, Length: 3},
		{Type: token.ARRAY_TYPE, Literal: "string", StartIndex: 11, Length: 6},
	}
	ctx22 := []token.TokenType{token.TYPE_NONE}

	input23 := "import ("
	expTokens23 := []token.Token{
		{Type: token.IMPORT, Literal: "import", StartIndex: 0, Length: 6},
		{Type: token.LPAREN, Literal: "(", StartIndex: 7, Length: 1},
	}
	ctx23 := []token.TokenType{token.TYPE_NONE, token.MULTI_IMPORT}

	input24 := `"os"`
	expTokens24 := []token.Token{
		{Type: token.DBL_QUOTE, Literal: "\"", StartIndex: 0, Length: 1},
		{Type: token.IMPORT_NAME, Literal: "os", StartIndex: 1, Length: 2},
		{Type: token.DBL_QUOTE, Literal: "\"", StartIndex: 3, Length: 1},
	}
	ctx24 := []token.TokenType{token.TYPE_NONE, token.MULTI_IMPORT}

	input25 := `display "github.com/cyamas/rizz/internal/display"`
	expTokens25 := []token.Token{
		{Type: token.IMPORT_ALIAS, Literal: "display", StartIndex: 0, Length: 7},
		{Type: token.DBL_QUOTE, Literal: "\"", StartIndex: 8, Length: 1},
		{Type: token.IMPORT_NAME, Literal: "github.com/cyamas/rizz/internal/display", StartIndex: 9, Length: 39},
		{Type: token.DBL_QUOTE, Literal: "\"", StartIndex: 48, Length: 1}}
	ctx25 := []token.TokenType{token.TYPE_NONE, token.MULTI_IMPORT}

	input26 := ")"
	expTokens26 := []token.Token{
		{Type: token.RPAREN, Literal: ")", StartIndex: 0, Length: 1},
	}
	ctx26 := []token.TokenType{token.TYPE_NONE, token.MULTI_IMPORT}

	tests := []struct {
		input      string
		context    []token.TokenType
		expTokens  []token.Token
		expContext token.TokenType
	}{
		{input1, ctx1, expTokens1, token.TYPE_NONE},
		{input2, ctx2, expTokens2, token.FUNC_BODY},
		{input3, ctx3, expTokens3, token.FUNC_BODY},
		{input4, ctx4, expTokens4, token.TYPE_NONE},
		{input5, ctx5, expTokens5, token.LOOP_BODY},
		{input6, ctx6, expTokens6, token.COND_BODY},
		{input7, ctx7, expTokens7, token.COND_BODY},
		{input8, ctx8, expTokens8, token.COND_BODY},
		{input9, ctx9, expTokens9, token.COND_BODY},
		{input10, ctx10, expTokens10, token.LOOP_BODY},
		{input11, ctx11, expTokens11, token.TYPE_NONE},
		{input12, ctx12, expTokens12, token.MAP_BODY},
		{input13, ctx13, expTokens13, token.MAP_BODY},
		{input14, ctx14, expTokens14, token.MAP_BODY},
		{input15, ctx15, expTokens15, token.TYPE_NONE},
		{input16, ctx16, expTokens16, token.TYPE_NONE},
		{input17, ctx17, expTokens17, token.TYPE_NONE},
		{input18, ctx18, expTokens18, token.TYPE_NONE},
		{input19, ctx19, expTokens19, token.TYPE_NONE},
		{input20, ctx20, expTokens20, token.TYPE_NONE},
		{input21, ctx21, expTokens21, token.TYPE_NONE},
		{input22, ctx22, expTokens22, token.TYPE_NONE},
		{input23, ctx23, expTokens23, token.MULTI_IMPORT},
		{input24, ctx24, expTokens24, token.MULTI_IMPORT},
		{input25, ctx25, expTokens25, token.MULTI_IMPORT},
		{input26, ctx26, expTokens26, token.TYPE_NONE},
	}

	var currTok token.Token
	for i, tt := range tests {
		h.ParseLine(tt.input, tt.context)

		for j, tok := range h.tokens {
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
			if tok.StartIndex != tt.expTokens[j].StartIndex {
				fmt.Printf("FAIL TEST %d\n", i+1)
				t.Fatalf("StartIndex should be %d. Got %d", tt.expTokens[j].StartIndex, tok.StartIndex)
			}
			if tok.Length != tt.expTokens[j].Length {
				fmt.Printf("FAIL TEST %d\n tokLiteral: %s, expLiteral: %s", i+1, tok.Literal, tt.expTokens[j].Literal)
				t.Fatalf("Length should be %d. Got %d", tt.expTokens[j].Length, tok.Length)
			}
		}
		if tt.expContext != h.context {
			fmt.Printf("FAIL TEST %d\n", i+1)
			fmt.Println("CURR TOK: ", currTok.Literal)
			fmt.Println("LEXER CONTEXT: ", l.Context())
			t.Fatalf("Context should be %q. Got %q", tt.expContext, h.context)
		}
	}
}
