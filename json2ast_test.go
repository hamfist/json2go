package json2go_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
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
		`{"a": 5, "b": true, "c": 64.5, "d": "foo"}`,
		`package main

		type Top struct {
		  a float64
          b bool
          c float64
          d string
		}`,
	},
	{
		`{"foo": {"bar": true}}`,
		`package main

		type Foo struct {
		  bar   bool
        }

		type Top struct {
		  foo	*Foo
		}`,
	},
	{
		`{"foo": [{"bar": true}, {"bar": false}]}`,
		`package main

		type Foo struct {
		  bar   bool
        }

		type Top struct {
		  foo	[]*Foo
		}`,
	},
	{
		`{"foo": [{"bar": true}, {"bar": false}]}`,
		`package main

		type Foo struct {
		  bar   bool
        }

		type Top struct {
		  foo	[]*Foo
		}`,
	},
	{
		`{"foo": [{"bar": {"foobar": 46.5}}]}`,
		`package main

        type Bar struct {
            foobar  float64
        }

		type Foo struct {
		  bar   *Bar
        }

		type Top struct {
		  foo	[]*Foo
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
	var obtainedNodes, expectedNodes chan *NodeWithBreadcrumbs
	var obtainedNode, expectedNode *NodeWithBreadcrumbs

	obtainedNodes = make(chan *NodeWithBreadcrumbs)
	expectedNodes = make(chan *NodeWithBreadcrumbs)

	go ast.Walk(&AstChannelWalker{Out: obtainedNodes}, params[0].(ast.Node))
	go ast.Walk(&AstChannelWalker{Out: expectedNodes}, params[1].(ast.Node))

	for {
		obtainedNode = <-obtainedNodes
		expectedNode = <-expectedNodes

		if obtainedNode == nil && expectedNode == nil {
			break
		}

		if obtainedNode == nil || expectedNode == nil {
			return false, fmt.Sprintf("\n%+v\ndid not match\n%+v", obtainedNode, expectedNode)
		}

		if !this.nodeEquals(obtainedNode.Node, expectedNode.Node) {
			return false, fmt.Sprintf("\n%+v\ndid not match\n%+v", obtainedNode, expectedNode)
		}
	}

	return true, ""
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

/* Partial AST comparison logic (for nodes and values we care about) */
type AstChannelWalker struct {
	Out         chan *NodeWithBreadcrumbs
	breadcrumbs Breadcrumbs
	walkCount   int
}

type Breadcrumbs []ast.Node

func (this Breadcrumbs) String() string {
	var buffer bytes.Buffer

	for _, node := range this {
		fmt.Fprintf(&buffer, " -> %+v", node)
	}

	return buffer.String()
}

type NodeWithBreadcrumbs struct {
	Node ast.Node
	Breadcrumbs
}

func (this *AstChannelWalker) Visit(node ast.Node) ast.Visitor {
	this.walkCount += 1

	if node != nil {
		this.Out <- &NodeWithBreadcrumbs{
			Node:        node,
			Breadcrumbs: this.breadcrumbs,
		}
		this.breadcrumbs = append(this.breadcrumbs, node)
	} else {
		this.breadcrumbs = this.breadcrumbs[0:len(this.breadcrumbs)]
		this.walkCount -= 2
	}

	//Walked all nodes
	if this.walkCount == 0 {
		close(this.Out)
	}

	return this
}

func valueEquals(obtained reflect.Value, expected reflect.Value) bool {
	if obtained.Kind() != expected.Kind() {
		return false
	}

	var result bool

	switch obtained.Kind() {
	case reflect.Bool:
		result = obtained.Bool() == expected.Bool()
	case reflect.String:
		result = obtained.String() == expected.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result = obtained.Int() == expected.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result = obtained.Uint() == expected.Uint()
	default:
		panic("Unknown kind: " + obtained.Kind().String())
	}

	return result
}

//Fields we care to compare in AST nodes
var fieldsForStruct = map[string][]string{
	"*ast.File":       []string{},
	"*ast.Ident":      []string{"Name"},
	"*ast.GenDecl":    []string{"Tok"},
	"*ast.TypeSpec":   []string{},
	"*ast.StructType": []string{"Incomplete"},
	"*ast.FieldList":  []string{},
	"*ast.Field":      []string{},
	"*ast.StarExpr":   []string{},
	"*ast.ArrayType":  []string{},
}

func (this *astEquals) nodeEquals(obtained ast.Node, expected ast.Node) bool {
	if reflect.TypeOf(obtained) != reflect.TypeOf(expected) {
		return false
	}
	var fields []string

	if fields = fieldsForStruct[reflect.TypeOf(obtained).String()]; fields == nil {
		return false
	}

	for _, field := range fields {
		if !valueEquals(reflect.Indirect(reflect.ValueOf(obtained)).FieldByName(field), reflect.Indirect(reflect.ValueOf(expected)).FieldByName(field)) {
			return false
		}
	}

	return true
}
