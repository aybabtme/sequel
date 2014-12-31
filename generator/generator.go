package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/aybabtme/sequel/reflector"
)

func Generate(w io.Writer, schema *reflector.DBSchema) error {
	bw := bufio.NewWriter(w)

	tmpl, err := template.New("db").Funcs(template.FuncMap{
		"camelize":        camelize,
		"pluralize":       pluralize,
		"var_to_go_type":  variableToGoType,
		"var_to_go_value": variableToGoValue,
		"col_to_go_type":  columnToGoType,
		"export":          export,
		"createQuery":     createQuery,
		"retrieveQuery":   retrieveQuery,
		"updateQuery":     updateQuery,
		"deleteQuery":     deleteQuery,
		"listQuery":       listQuery,
		"listIndex":       listIndex,
	}).Parse(clientTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(bw, schema)
	if err != nil {
		return err
	}

	return bw.Flush()
}

func camelize(str string) string {

	converts := map[string]string{
		"_id": "_ID",
	}

	b := []byte(str)

	for from, to := range converts {
		b = bytes.Replace(b, []byte(from), []byte(to), -1)
	}

	strs := bytes.Split(b, []byte("_"))

	buf := bytes.NewBuffer([]byte(strs[0]))
	for i, val := range strs {
		if i == 0 {
			continue
		}
		if len(val) < 1 {
			continue
		}
		buf.Write(bytes.Title(val))
	}

	return buf.String()
}

func export(str string) string {
	if len(str) < 1 {
		return str
	}

	if unicode.IsLower([]rune(str)[0]) {
		return strings.Title(str)
	}

	return str
}

func pluralize(str string) string {
	if str[len(str)-1] != 's' {
		return str + "s"
	}

	return str
}

func variableToGoType(v reflector.Variable) string {
	switch v.Type {
	case reflector.SQLString:
		return "string"
	case reflector.SQLBytes:
		return "[]byte"
	case reflector.SQLInteger:
		return "int"
	case reflector.SQLFloat:
		return "float64"
	case reflector.SQLBool:
		return "bool"
	case reflector.SQLTime:
		return "time.Time"
	}
	panic(v)
}

func variableToGoValue(v reflector.Variable) string {
	return fmt.Sprintf("%#v", v.Value)
}

func columnToGoType(c reflector.Column) string {
	if c.Nullable {
		switch c.Type {
		case reflector.SQLString:
			return "NullString"
		case reflector.SQLBytes:
			return "[]byte"
		case reflector.SQLInteger:
			return "NullInt64"
		case reflector.SQLFloat:
			return "NullFloat64"
		case reflector.SQLBool:
			return "NullBool"
		case reflector.SQLTime:
			return "NullTime"
		}
	}

	switch c.Type {
	case reflector.SQLString:
		return "string"
	case reflector.SQLBytes:
		return "[]byte"
	case reflector.SQLInteger:
		return "int"
	case reflector.SQLFloat:
		return "float64"
	case reflector.SQLBool:
		return "bool"
	case reflector.SQLTime:
		return "time.Time"
	}
	panic(c)
}
