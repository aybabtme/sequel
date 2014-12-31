package generator_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"

	"github.com/aybabtme/sequel/generator"
	"github.com/aybabtme/sequel/reflector"

	_ "github.com/go-sql-driver/mysql"
)

const dsn = "root@tcp(127.0.0.1:3306)/test_alpha_b9cc2531c0"

func TestCanGenerateClient(t *testing.T) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	schema, err := reflector.DescribeMySQL(db, "test_alpha_b9cc2531c0")
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer(nil)
	err = generator.Generate(buf, schema)
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, "", buf, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	out := bytes.Buffer{}

	err = printer.Fprint(&out, fset, ast)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out.String())
}
