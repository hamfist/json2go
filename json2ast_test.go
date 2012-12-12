package json2go_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"regexp"
	"testing"
)

import (
	. "launchpad.net/gocheck"
)

import (
	"github.com/modcloth/json2go"
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
		`package main

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

//Poor man's AST comparison that ignores whitespace. TODO Revisit.
var WhitespaceRegexp = regexp.MustCompile(`\s+`)

func (this *astEquals) Check(params []interface{}, names []string) (bool, string) {
	var obtainedAstString, expectedAstString *bytes.Buffer

	obtainedAstString = new(bytes.Buffer)
	expectedAstString = new(bytes.Buffer)

	printer.Fprint(obtainedAstString, token.NewFileSet(), params[0])
	printer.Fprint(expectedAstString, token.NewFileSet(), params[1])

	//Couldn't replace with one space because the first example output
	//loses the space inbetween the struct keyword and { when parsed.
	//TODO Determine if this is a bug in the AST parser/printer
	trimmedObtained := WhitespaceRegexp.ReplaceAll(obtainedAstString.Bytes(), []byte(""))
	trimmedExpected := WhitespaceRegexp.ReplaceAll(expectedAstString.Bytes(), []byte(""))

	if string(trimmedObtained) == string(trimmedExpected) {
		return true, ""
	}

	return false, fmt.Sprintf("\n%s\ndid not match\n%s", trimmedObtained, trimmedExpected)
}

func (s *Json2AstSuite) TestJson2Ast(c *C) {
	var fset *token.FileSet
	var obtainedAst, expectedAst *ast.File
	var err error

	fset = token.NewFileSet()

	for i, tt := range json2AstTests {
		if expectedAst, err = parser.ParseFile(fset, "", tt.Output, 0); err != nil {
			c.Error(err)
			continue
		}

		if obtainedAst, err = json2go.Json2Ast(bytes.NewBufferString(tt.Input)); err != nil {
			c.Error(err)
			continue
		}

		c.Assert(obtainedAst, AstEquals, expectedAst, Commentf("Test Case %d failed", i))
	}
}
