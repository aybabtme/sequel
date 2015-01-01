package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/aybabtme/sequel/generator/tmpl"
	"github.com/aybabtme/sequel/reflector"
)

var (
	DefaultDirperm  os.FileMode = 0755
	DefaultFileperm os.FileMode = 0644
)

func Generate(dirname string, schema *reflector.DBSchema) error {

	files := map[string][]byte{}
	t := template.New("root").Funcs(funcmap)

	compileAndAdd := func(tname, tcontent string, tvalue interface{}) error {
		buf := bytes.NewBuffer(nil)
		subT, err := t.New(tname).Parse(tcontent)
		if err != nil {
			return err
		}
		if err := subT.Execute(buf, tvalue); err != nil {
			return err
		}
		gofmted, err := gofmt(buf.Bytes())
		if err != nil {
			return fmt.Errorf("bad template %q: %v", tname, err)
		}
		files[tname] = gofmted
		return nil
	}

	for tname, tcontent := range map[string]string{
		"client.go":      tmpl.ClientTemplate,
		"client_test.go": tmpl.ClientTestTemplate,
		"common.go":      tmpl.CommonTemplate,
		"common_test.go": tmpl.CommonTestTemplate,
	} {
		err := compileAndAdd(tname, tcontent, schema)
		if err != nil {
			return err
		}
	}

	needsTime := func(tbl reflector.Table) bool {
		if tbl.Has("created_at") != nil || tbl.Has("updated_at") != nil {
			return true
		}
		for _, col := range tbl.Columns {
			if col.Type == reflector.SQLTime && !col.Nullable {
				return true
			}
		}
		return false
	}

	for _, tbl := range schema.Tables {
		basename := pluralize(tbl.Name)
		filename := basename + ".go"
		testFilename := basename + "_test.go"

		tval := map[string]interface{}{
			"DB":           schema,
			"Tbl":          tbl,
			"HasCreatedAt": tbl.Has("created_at"),
			"HasUpdatedAt": tbl.Has("updated_at"),
			"NeedsTime":    needsTime(tbl),
		}

		for tname, tcontent := range map[string]string{
			filename:     tmpl.TableTemplate,
			testFilename: tmpl.TableTestTemplate,
		} {
			err := compileAndAdd(tname, tcontent, tval)
			if err != nil {
				return err
			}
		}
	}

	return createFiles(filepath.Join(dirname, schema.Name), files)
}

func createFiles(dirname string, files map[string][]byte) error {

	if err := os.MkdirAll(dirname, DefaultDirperm); err != nil {
		return err
	}

	fname := func(name string) string {
		return filepath.Join(dirname, name)
	}

	for name, content := range files {

		err := ioutil.WriteFile(
			fname(name),
			content,
			DefaultFileperm,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func gofmt(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	out := bytes.Buffer{}

	err = printer.Fprint(&out, fset, ast)
	if err != nil {
		return nil, err
	}

	fmted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, err
	}

	return fmted, nil
}
