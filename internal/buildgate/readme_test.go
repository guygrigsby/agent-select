package buildgate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeRoutingSurface(t *testing.T) {
	root := filepath.Join("..", "..")
	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "routing", "route.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	implemented := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Route" || fn.Recv == nil || fn.Body == nil {
			continue
		}
		ptr, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
		if ok {
			id, ok := ptr.X.(*ast.Ident)
			implemented = ok && id.Name == "Engine"
		}
	}
	if implemented != strings.Contains(string(readme), "`Engine.Route`") || strings.Contains(string(readme), "No router or collection service is implemented") {
		t.Fatal("remedy: update README.md routing scope to match the implemented Engine.Route surface")
	}
}
