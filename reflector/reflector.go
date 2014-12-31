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
	Name   string
	Tables []Table
}

func DescribeMySQL(db Queryer, dbname string) (*DBSchema, error) {

	schema := &DBSchema{
		Name: dbname,
	}
	return schema, schema.load(db)
}

func (db *DBSchema) load(q Queryer) error {
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
	fmt.Fprintf(buf, "db %q, %d tables\n", db.Name, len(db.Tables))
	for i, tbl := range db.Tables {
		fmt.Fprintf(buf, "\t%d: %s\n", i, tbl.String())
	}
	return buf.String()
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
	Key     sql.RawBytes
	Default sql.RawBytes
	Extra   sql.RawBytes
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
	col.Type, err = parseSQLType(typeName)
	return err
}

func (col Column) String() string {
	buf := bytes.NewBuffer(nil)

	fmt.Fprintf(buf, "%s \t", col.Name)
	if col.Nullable {
		fmt.Fprintf(buf, "*")
	}
	fmt.Fprintf(buf, "%s", col.Type)

	if len(col.Default) != 0 {
		fmt.Fprintf(buf, " = %q", string(col.Default))
	}
	if len(col.Key) != 0 {
		fmt.Fprintf(buf, " \t// key %q", string(col.Key))
	}
	if len(col.Extra) != 0 {
		fmt.Fprintf(buf, ", extra %q", string(col.Extra))
	}

	return buf.String()
}
