package token

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT            = "IDENT"
	INT_LITERAL      = "INTLITERAL"
	ASSIGN           = "="
	SHORT_VAR_ASSIGN = ":="
	PLUS             = "+"
	MINUS            = "-"
	BANG             = "!"
	ASTERISK         = "*"
	SLASH            = "/"

	LT = "<"
	GT = ">"
	EQ = "=="

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	DBL_QUOTE = "\""
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"

	//keywords
	FUNC_DECLARE = "FUNCDECLARE"
	RETURN       = "RETURN"
	VAR_DECLARE  = "VARDECLARE"
	TYPE         = "TYPE"
	FOR          = "FOR"
	RANGE        = "RANGE"
	IF           = "IF"
	ELSE         = "ELSE"
	TRUE         = "TRUE"
	FALSE        = "FALSE"

	//types
	INT       = "INT"
	FLOAT_32  = "FLOAT32"
	FLOAT_64  = "FLOAT64"
	BOOL      = "BOOL"
	STRUCT    = "STRUCT"
	INTERFACE = "INTERFACE"
	SLICE     = "TYPESLICE"
	STRING    = "TYPESTRING"
	TYPE_NONE = "TYPENONE"

	MAP_DECLARE = "MAPDECLARE"
	MAP_SIG     = "MAPSIG"
	MAP_BODY    = "MAPBODY"
	KEY         = "KEY"
	VALUE       = "VALUE"

	FIELD_NAME = "FIELDNAME"

	STRING_LITERAL = "STRINGLITERAL"
	VAR_NAME       = "VARNAME"

	FUNC_SIG     = "FUNCSIG"
	FUNC_NAME    = "FUNCNAME"
	PARAM_NAME   = "PARAMNAME"
	PARAM_TYPE   = "PARAM_TYPE"
	START_PARAMS = "STARTPARAMS"
	END_PARAMS   = "END_PARAMS"
	RETURN_TYPE  = "RETURNTYPE"
	FUNC_BODY    = "FUNC_BODY"

	FOR_SIG   = "FORSIG"
	LOOP_BODY = "LOOPBODY"

	IF_SIG    = "IFSIG"
	COND_BODY = "COND_BODY"
)

var keywords = map[string]TokenType{
	"func":      FUNC_DECLARE,
	"var":       VAR_DECLARE,
	"int":       INT,
	"float32":   FLOAT_32,
	"float64":   FLOAT_64,
	"string":    STRING,
	"return":    RETURN,
	"type":      TYPE,
	"struct":    STRUCT,
	"interface": INTERFACE,
	"map":       MAP_DECLARE,
	"for":       FOR,
	"range":     RANGE,
	"if":        IF,
	"else":      ELSE,
	"true":      TRUE,
	"false":     FALSE,
	"bool":      BOOL,
}

func LookupIdent(ident string, context TokenType) TokenType {
	if tok, ok := keywords[ident]; ok {
		switch context {
		case PARAM_NAME:
			return PARAM_TYPE
		case END_PARAMS:
			return RETURN_TYPE
		}
		return tok
	}
	switch context {
	case DBL_QUOTE:
		return STRING_LITERAL
	case VAR_DECLARE:
		return VAR_NAME
	case FUNC_DECLARE:
		return FUNC_NAME
	case START_PARAMS:
		return PARAM_NAME
	case STRUCT:
		return FIELD_NAME
	case MAP_BODY:
		return KEY
	case KEY:
		return VALUE
	}
	return IDENT
}

type TokenType string

type Token struct {
	Type       TokenType
	Literal    string
	startIndex int
	length     int
}

func (t *Token) Length() int {
	return t.length
}

func (t *Token) SetLength(l int) {
	t.length = l
}

func (t *Token) SetIndex(idx int) {
	t.startIndex = idx
}

func (t *Token) StartIndex() int {
	return t.startIndex
}
