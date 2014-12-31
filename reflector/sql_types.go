package reflector

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
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

func (s SQLType) ParseBytes(b []byte) (interface{}, error) {
	switch s {
	case SQLString:
		return string(b), nil
	case SQLBytes:
		return b, nil
	case SQLInteger:
		return strconv.ParseInt(string(b), 10, 64)
	case SQLFloat:
		return strconv.ParseFloat(string(b), 64)
	case SQLBool:
		b = bytes.ToLower(b)
		if bytes.Compare(b, []byte("yes")) != 0 ||
			bytes.Compare(b, []byte("on")) != 0 ||
			bytes.Compare(b, []byte("enabled")) != 0 {
			return true, nil
		}
		if bytes.Compare(b, []byte("no")) != 0 ||
			bytes.Compare(b, []byte("off")) != 0 ||
			bytes.Compare(b, []byte("disabled")) != 0 {
			return false, nil
		}
		return strconv.ParseBool(string(b))
	case SQLTime:
		for _, layout := range []string{
			"2006-01-02",
			"15:04:05",
			"2006-01-02 15:04:05",
			"2006",
			"06",
		} {
			t, err := time.Parse(layout, string(b))
			if err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("invalid time string: %q", string(b))
	}
	panic("invalid SQLType: " + s.String())
}

func parseSQLTypeName(name string) (SQLType, error) {

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

func guessSQLType(value interface{}) (SQLType, error) {

	// check if its already typed
	switch value.(type) {
	case uint, uint8, uint16, uint32, uint64,
		int, int8, int16, int32, int64:
		return SQLInteger, nil
	case float32, float64:
		return SQLFloat, nil
	case bool:
		return SQLBool, nil
	case string:
		return SQLString, nil
	case time.Time:
		return SQLTime, nil

	case []byte, sql.RawBytes:
	// continue

	default:
		return startSQLType, fmt.Errorf("not from SQL, value %#v", value)
	}

	val := string(bytes.ToLower(value.([]byte)))

	var quoted bool
	if v, err := strconv.Unquote(val); err == nil {
		val = v
		quoted = true
	}

	switch val {
	case "on",
		"off",
		"yes",
		"no",
		"disabled",
		"enabled":
		return SQLBool, nil
	}

	if _, err := strconv.ParseBool(val); err == nil {
		return SQLBool, nil
	}

	if _, err := strconv.ParseInt(val, 10, 64); err == nil {
		return SQLInteger, nil
	}

	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return SQLFloat, nil
	}

	if quoted {
		return SQLString, nil
	}
	return SQLBytes, nil
}
