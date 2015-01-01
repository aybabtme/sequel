package generator_test

import (
	"database/sql"
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
	err = generator.Generate("test_pkg", schema)
	if err != nil {
		t.Fatal(err)
	}
}
