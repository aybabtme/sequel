package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	"github.com/aybabtme/sequel/reflector"
)

var funcmap = template.FuncMap{
	"camelize":           camelize,
	"explode_underscore": explode_underscore,
	"singularize":        singularize,
	"pluralize":          pluralize,
	"var_to_go_type":     variableToGoType,
	"var_to_go_value":    variableToGoValue,
	"col_to_go_type":     columnToGoType,
	"idx_list_args":      idxListArgs,
	"idx_query_args":     idxQueryArgs,
	"export":             export,

	"createQuery":   createQuery,
	"retrieveQuery": retrieveQuery,
	"updateQuery":   updateQuery,
	"deleteQuery":   deleteQuery,
	"listQuery":     listQuery,
	"listIndex":     listIndex,
}

func camelize(str string) string {

	if str == "id" {
		return "ID"
	}

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

func explode_underscore(str string) string {
	return strings.Join(strings.Split(str, "_"), " ")
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

func idxListArgs(idx reflector.Index) string {
	buf := bytes.NewBuffer(nil)
	for i, col := range idx.Columns {
		if i != 0 {
			fmt.Fprint(buf, ", ")
		}

		t := columnToGoType(col)
		if i+1 < len(idx.Columns) && t == columnToGoType(idx.Columns[i+1]) {
			fmt.Fprintf(buf, "%s", camelize(col.Name))
		} else {
			fmt.Fprintf(buf, "%s %s", camelize(col.Name), t)
		}

	}
	return buf.String()
}
func idxQueryArgs(idx reflector.Index) string {
	buf := bytes.NewBuffer(nil)
	for i, col := range idx.Columns {
		if i != 0 {
			fmt.Fprint(buf, ", ")
		}
		fmt.Fprintf(buf, "%s", camelize(col.Name))
	}
	return buf.String()
}

var singularToPlurar = map[string]string{
	"reply": "replies",
}

func pluralize(str string) string {
	for sing, plur := range singularToPlurar {
		str = strings.Replace(str, sing, plur, -1)
		str = strings.Replace(str, strings.Title(sing), strings.Title(plur), -1)
		str = strings.Replace(str, strings.ToUpper(sing), strings.ToUpper(plur), -1)
		str = strings.Replace(str, strings.ToLower(sing), strings.ToLower(plur), -1)
	}

	if str[len(str)-1] != 's' {
		return str + "s"
	}

	return str
}

func singularize(str string) string {
	for sing, plur := range singularToPlurar {
		str = strings.Replace(str, plur, sing, -1)
		str = strings.Replace(str, strings.Title(plur), strings.Title(sing), -1)
		str = strings.Replace(str, strings.ToUpper(plur), strings.ToUpper(sing), -1)
		str = strings.Replace(str, strings.ToLower(plur), strings.ToLower(sing), -1)
	}

	if str[len(str)-1] == 's' {
		return str[:len(str)-1]
	}
	return str
}
