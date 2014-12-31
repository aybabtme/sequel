package reflector

import (
	"bytes"
	"database/sql"
	"fmt"
	"text/tabwriter"
)

type Queryer interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}

/*
Database!
*/

type DBSchema struct {
	Name      string
	Variables []Variable
	Tables    []Table
}

func DescribeMySQL(db Queryer, dbname string) (*DBSchema, error) {

	schema := &DBSchema{
		Name: dbname,
	}

	return schema, schema.load(db)
}

func (db *DBSchema) load(q Queryer) error {
	if err := db.loadVariables(q); err != nil {
		return err
	}
	return db.loadTables(q)
}

func (db *DBSchema) loadVariables(q Queryer) error {

	rows, err := q.Query("show variables")
	if err != nil {
		return fmt.Errorf("showing variables, %v", err)
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		v := Variable{}
		if err := v.scan(rows); err != nil {
			return fmt.Errorf("scanning variable %d, %v", i, err)
		}
		db.Variables = append(db.Variables, v)
	}

	return rows.Err()
}

func (db *DBSchema) loadTables(q Queryer) error {

	rows, err := q.Query("show tables")
	if err != nil {
		return fmt.Errorf("showing tables, %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		tbl := Table{}
		if err := rows.Scan(&tbl.Name); err != nil {
			return err
		}
		if err := tbl.load(q); err != nil {
			return fmt.Errorf("loading table %q, %v", tbl.Name, err)
		}
		db.Tables = append(db.Tables, tbl)
	}

	return rows.Err()
}

func (db DBSchema) String() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "db %q, %d variables, %d tables\n", db.Name, len(db.Variables), len(db.Tables))

	w := tabwriter.NewWriter(buf, 8, 8, 0, ' ', 0)

	fmt.Fprintf(w, "var (\n")
	for _, v := range db.Variables {
		fmt.Fprintf(w, "\t%s\n", v.String())
	}
	fmt.Fprintf(w, ")\n")
	w.Flush()

	for i, tbl := range db.Tables {
		fmt.Fprintf(buf, "\t%d: %s\n", i, tbl.String())
	}
	return buf.String()
}

/*
Variables!
*/

type Variable struct {
	Name  string
	Type  SQLType
	Value interface{}
}

func (v *Variable) scan(rows *sql.Rows) error {
	var values []byte
	err := rows.Scan(&v.Name, &values)
	if err != nil {
		return err
	}
	v.Type, err = guessSQLType(values)
	if err != nil {
		return err
	}

	v.Value, err = v.Type.ParseBytes(values)

	return err
}

func (v *Variable) String() string {
	switch val := v.Value.(type) {
	case []byte:
		return fmt.Sprintf("%s \t%#v", v.Name, string(val))
	}

	return fmt.Sprintf("%s \t%#v", v.Name, v.Value)
}

/*
Tables!
*/

type Table struct {
	Name    string
	Columns []Column
	// add more stuff like keys
}

func (tbl *Table) load(q Queryer) error {
	// sprintf'ing queries, because yolo (because prepared stmts dont
	// work for dtl)
	rows, err := q.Query(fmt.Sprintf("describe %s", tbl.Name))
	if err != nil {
		return fmt.Errorf("describing table %q, %v", tbl.Name, err)
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		col := Column{}
		err := col.scan(rows)
		if err != nil {
			return fmt.Errorf("scanning column %d, %v", i, err)
		}
		tbl.Columns = append(tbl.Columns, col)
	}

	return rows.Err()
}

func (tbl Table) String() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "\ttable %q, %d columns\n", tbl.Name, len(tbl.Columns))

	w := tabwriter.NewWriter(buf, 8, 8, 0, ' ', 0)

	fmt.Fprintf(w, "\tvar (\n")
	for _, col := range tbl.Columns {
		fmt.Fprintf(w, "\t\t%s\n", col.String())
	}
	fmt.Fprintf(w, "\t)\n")
	w.Flush()
	return buf.String()
}

/*
Columns!
*/

type Column struct {
	Name     string
	Type     SQLType
	Nullable bool
	// add more stuff like `key` and `extra`
	Key     interface{}
	Default interface{}
	Extra   interface{}
}

func (col *Column) scan(rows *sql.Rows) error {
	var typeName string
	var nullable string
	err := rows.Scan(
		&col.Name,
		&typeName,
		&nullable,
		&col.Key,
		&col.Default,
		&col.Extra,
	)
	if err != nil {
		return err
	}
	col.Nullable = ("YES" == nullable)
	col.Type, err = parseSQLTypeName(typeName)
	if err != nil {
		return err
	}
	if def, ok := col.Default.([]byte); ok {
		col.Default, err = col.Type.ParseBytes(def)
	}
	return err
}

func (col Column) String() string {
	buf := bytes.NewBuffer(nil)

	fmt.Fprintf(buf, "%s \t", col.Name)
	if col.Nullable {
		fmt.Fprintf(buf, "*")
	}
	fmt.Fprintf(buf, "%s", col.Type)

	if col.Default != nil {
		fmt.Fprintf(buf, " = %#v", col.Default)
	}
	var keyed bool
	if key, ok := col.Key.([]byte); ok && len(key) != 0 {
		fmt.Fprintf(buf, " \t// key %q", string(key))
		keyed = true
	}
	if extra, ok := col.Extra.([]byte); ok && len(extra) != 0 {
		if !keyed {
			fmt.Fprintf(buf, "\t//")
		}
		fmt.Fprintf(buf, " extra %q", string(extra))
	}

	return buf.String()
}
