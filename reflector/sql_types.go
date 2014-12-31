package reflector

import (
	"fmt"
	"strings"
)

type SQLType int

const (
	startSQLType SQLType = iota

	SQLString
	SQLBytes
	SQLInteger
	SQLFloat
	SQLBool
	SQLTime

	stopSQLType
)

func parseSQLType(name string) (SQLType, error) {

	if l := strings.Index(name, "("); l > 0 && l < len(name) {
		name = name[:l]
	}
	name = strings.ToLower(name)

	var (
		t   SQLType
		err error
	)
	switch name {

	case "integer", "serial", "tinybit", "bit",
		"tinyint", "smallint", "mediumint", "int", "bigint":
		t = SQLInteger

	case "decimal", "dec", "fixed", "numeric", "float", "double",
		"double precision":
		t = SQLFloat

	case "bool", "boolean":
		t = SQLBool

	case "datetime", "date", "timestamp", "time", "year":
		t = SQLTime

	case "char", "varchar", "character",
		"tinytext", "mediumtext", "text", "longtext":
		t = SQLString

	case "char byte", "binary", "varbinary",
		"tinyblob", "mediumblob", "blob", "longblob":
		t = SQLBytes

	default:
		err = fmt.Errorf("unknown type %q", name)
	}
	return t, err
}
