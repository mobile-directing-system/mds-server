package testutil

import (
	"fmt"
	"github.com/lefinal/nulls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

// ExtractStringConstantsOfType extracts all constant values for constants with
// the given type and prefix in the file.
func ExtractStringConstantsOfType[T ~string](path string, constPrefix string) ([]T, error) {
	constants := make([]T, 0)
	// Extract constants.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, token.LowestPrec)
	if err != nil {
		return nil, err
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if len(valueSpec.Names) != 1 {
				continue
			}
			if valueName := valueSpec.Names[0].Name; !strings.HasPrefix(valueName, constPrefix) {
				continue
			}
			valueSpecType, ok := valueSpec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			if valueSpecType.Name != reflect.TypeOf(T("")).Name() {
				continue
			}
			if len(valueSpec.Values) != 1 {
				continue
			}
			valueSpecValue, ok := valueSpec.Values[0].(*ast.BasicLit)
			if !ok {
				continue
			}
			v := valueSpecValue.Value
			v = strings.Trim(v, `"`)
			constants = append(constants, T(v))
		}
	}
	return constants, nil
}

// TestMapperWithConstExtraction tests the given mapper function with constants
// from the given file. All constants are extracted using
// ExtractStringConstantsOfType and then assured that mapping does not fail.
// Finally, an unknown constant is used which is assured to make the mapper
// function return an error.
func TestMapperWithConstExtraction[From ~string, To ~string](t *testing.T, mapperFn func(From) (To, error),
	filepath string, prefixOverwrite nulls.String) {
	prefix := reflect.TypeOf(From("")).Name()
	if prefixOverwrite.Valid {
		prefix = prefixOverwrite.String
	}
	constants, err := ExtractStringConstantsOfType[From](filepath, prefix)
	require.NoError(t, err, "extracting constants should not fail")
	require.NotEmptyf(t, constants, "should have found types (searching for type '%s' with prefix '%s')",
		reflect.TypeOf(From("")).Name(), prefix)

	for _, c := range constants {
		c := c
		t.Run(fmt.Sprintf("TestConstant_%s", c), func(t *testing.T) {
			_, err := mapperFn(c)
			assert.NoErrorf(t, err, "mapping should not fail for type %v", c)
		})
	}
	t.Run("TestUnknownConstant", func(t *testing.T) {
		_, err = mapperFn(From(NewUUIDV4().String()))
		assert.Error(t, err, "should fail for unknown type")
	})
}
