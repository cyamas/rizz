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

	SINGLE_QUOTE = "'"
	DBL_QUOTE    = "\""
	LPAREN       = "("
	RPAREN       = ")"
	LBRACE       = "{"
	RBRACE       = "}"
	LBRACKET     = "["
	RBRACKET     = "]"

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
	ANY          = "ANY"
	MAKE         = "MAKE"
	IMPORT       = "IMPORT"
	PACKAGE      = "PACKAGE"
	MAIN         = "MAIN"
	LEN          = "LEN"

	PACKAGE_NAME = "PACKAGENAME"

	IMPORT_NAME         = "IMPORTNAME"
	MULTI_IMPORT        = "MULTIIMPORT"
	SINGLE_IMPORT       = "SINGLEIMPORT"
	START_SINGLE_IMPORT = "STARTSINGLEIMPORT"
	START_IMPORT_NAME   = "STARTIMPORTNAME"
	IMPORT_ALIAS        = "IMPORTALIAS"
	IMPORT_CALL         = "IMPORTCALL"

	//types
	INT       = "INT"
	UINT_8    = "UINT8"
	FLOAT_32  = "FLOAT32"
	FLOAT_64  = "FLOAT64"
	BOOL      = "BOOL"
	STRUCT    = "STRUCT"
	RUNE      = "RUNE"
	INTERFACE = "INTERFACE"
	STRING    = "TYPESTRING"
	BYTE      = "BYTE"
	TYPE_NONE = "TYPENONE"

	TYPE_NAME = "TYPENAME"
	TYPE_TYPE = "TYPE_TYPE"
	TYPE_CALL = "TYPECALL"

	MAP_DECLARE = "MAPDECLARE"
	KEY_DECLARE = "KEYDECLARE"
	VAL_DECLARE = "VALDECLARE"
	MAP_SIG     = "MAPSIG"
	MAP_BODY    = "MAPBODY"
	KEY         = "KEY"
	VALUE       = "VALUE"

	SLICE_DECLARE = "SLICEDECLARE"
	SLICE_BODY    = "SLICEBODY"
	ITEM_TYPE     = "ITEMTYPE"
	ITEM          = "ITEM"

	ARRAY_DECLARE = "ARRAYDECLARE"
	ARRAY_TYPE    = "ARRAYTYPE"

	FIELD_NAME = "FIELDNAME"

	RUNE_LITERAL   = "RUNELITERAL"
	STRING_LITERAL = "STRINGLITERAL"
	FLOAT_LITERAL  = "FLOATLITERAL"
	VAR_NAME       = "VARNAME"
	VAR_CALL       = "VARCALL"

	RUNE_START = "RUNESTART"

	FUNC_SIG     = "FUNCSIG"
	FUNC_NAME    = "FUNCNAME"
	PARAM_NAME   = "PARAMNAME"
	PARAM_TYPE   = "PARAM_TYPE"
	START_PARAMS = "STARTPARAMS"
	END_PARAMS   = "END_PARAMS"
	RETURN_TYPE  = "RETURNTYPE"
	FUNC_BODY    = "FUNC_BODY"
	FUNC_CALL    = "FUNCCALL"

	FOR_SIG   = "FORSIG"
	LOOP_BODY = "LOOPBODY"

	IF_SIG    = "IFSIG"
	COND_BODY = "COND_BODY"
)

var keywords = map[string]TokenType{
	"package":   PACKAGE,
	"import":    IMPORT,
	"func":      FUNC_DECLARE,
	"var":       VAR_DECLARE,
	"int":       INT,
	"uint8":     UINT_8,
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
	"rune":      RUNE,
	"byte":      BYTE,
	"any":       ANY,
	"make":      MAKE,
	"main":      MAIN,
	"len":       LEN,
}

func LookupIdent(ident string, context TokenType) TokenType {
	if tok, ok := keywords[ident]; ok {
		switch context {
		case PACKAGE:
			return PACKAGE_NAME
		case DBL_QUOTE:
			return STRING_LITERAL
		case SINGLE_QUOTE:
			return RUNE_LITERAL
		case ARRAY_DECLARE:
			return ARRAY_TYPE
		case SLICE_DECLARE:
			return ITEM_TYPE
		case KEY:
			return VALUE
		case MAP_SIG:
			return KEY_DECLARE
		case KEY_DECLARE:
			return VAL_DECLARE
		case PARAM_NAME:
			return PARAM_TYPE
		case END_PARAMS:
			return RETURN_TYPE
		}
		return tok
	}
	switch context {
	case SINGLE_IMPORT:
		return IMPORT_ALIAS
	case MULTI_IMPORT:
		return IMPORT_ALIAS
	case TYPE:
		return TYPE_NAME
	case START_IMPORT_NAME:
		return IMPORT_NAME
	case RUNE_START:
		return RUNE_LITERAL
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
	StartIndex int
	Length     int
}

func (t *Token) GetLength() int {
	return t.Length
}

func (t *Token) SetLength(l int) {
	t.Length = l
}

func (t *Token) SetIndex(idx int) {
	t.StartIndex = idx
}

func (t *Token) GetStartIndex() int {
	return t.StartIndex
}
