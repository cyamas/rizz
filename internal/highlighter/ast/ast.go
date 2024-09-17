package ast

import "github.com/cyamas/rizz/internal/highlighter/token"

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

type VarDeclareStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (vds *VarDeclareStatement) statementNode()       {}
func (vds *VarDeclareStatement) TokenLiteral() string { return vds.Token.Literal }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
