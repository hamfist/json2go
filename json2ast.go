package json2go

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
	"io"
)

func convertMap(name string, v map[string]interface{}) ([]ast.Decl, error) {
	var err error
	var fields []*ast.Field
	var nestedDecls []ast.Decl

	output := make([]ast.Decl, 0)

	for key := range v {
		var t ast.Expr
		valueType := reflect.TypeOf(v[key])
		switch valueType.Kind() {
		case reflect.Map:
			tName := strings.Title(key)
			if nestedDecls, err = convertMap(tName, v[key].(map[string]interface{})); err != nil {
				return nil, err
			}

			for _, nestedDecl := range nestedDecls {
				output = append(output, nestedDecl)
			}
			t = &ast.StarExpr{
				X: ast.NewIdent(tName),
			}
		case reflect.Slice:
			slice := v[key].([]interface{})
			if len(slice) == 0 {
				t = &ast.InterfaceType{
					Methods: &ast.FieldList{},
				}
				break
			}
			sliceType := reflect.TypeOf(slice[0])
			switch sliceType.Kind() {
			case reflect.Map:
				tName := strings.Title(key)
				if nestedDecls, err = convertMap(tName, slice[0].(map[string]interface{})); err != nil {
					return nil, err
				}

				for _, nestedDecl := range nestedDecls {
					output = append(output, nestedDecl)
				}
				t = &ast.StarExpr{
					X: ast.NewIdent(tName),
				}
			default:
				t = ast.NewIdent(sliceType.String())
			}

			t = &ast.ArrayType{
				Elt: t,
			}
		default:
			t = ast.NewIdent(reflect.TypeOf(v[key]).String())
		}

		fields = append(fields, &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name:    key,
					NamePos: token.NoPos,
					Obj:     ast.NewObj(ast.Var, key),
				},
			},
			Type: t,
		})
	}

	output = append(output, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	})

	return output, err
}

func Json2Ast(jsonReader io.Reader) (*ast.File, error) {
	var v map[string]interface{}
	var file *ast.File
	var err error

	dec := json.NewDecoder(jsonReader)

	file = &ast.File{
		Name: ast.NewIdent("main"),
	}

	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	fmt.Printf("%s\n", reflect.TypeOf(v["bar"]))

	var decls []ast.Decl
	if decls, err = convertMap("Top", v); err != nil {
	  return nil, err
	}

	file.Decls = decls

	return file, nil
}
