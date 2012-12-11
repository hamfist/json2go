package json2go

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

import (
	. "launchpad.net/gocheck"
)

func Test(t *testing.T) { TestingT(t) }

type Json2AstSuite struct{}

var _ = Suite(&Json2AstSuite{})

var json2AstTests = []struct {
	Input  string
	Output string
}{
	{
		`{"foo": 5}`,
		`
		package main

		type Top struct {
		  foo	float64
		}`,
	},
}

type astEquals struct {
	*CheckerInfo
}

var AstEquals Checker = &astEquals{
	&CheckerInfo{Name: "AST Equals", Params: []string{"obtained", "expected"}},
}

func (this *astEquals) Check(params []interface{}, names []string) (bool, string) {
	return false, ""
}

func (s *Json2AstSuite) TestJson2Ast(c *C) {
	var fset *token.FileSet
	var structAst *ast.File
	var err error

	fset = token.NewFileSet()

	for _, tt := range json2AstTests {
		if structAst, err = parser.ParseFile(fset, "", tt.Output, 0); err != nil {
			panic(err)
		}

		ast.Print(fset, structAst)
		//c.Assert(tt.
		//c.Assert(lucasResult.Type, Equals, tt.lucasSendResult.Type, Commentf("Test Case %d failed", i))
	}
}
