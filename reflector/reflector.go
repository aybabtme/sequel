package reflector

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

type queryer interface {
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

func DescribeMySQL(db *sql.DB, dbname string) (*DBSchema, error) {

	schema := &DBSchema{
		Name: dbname,
	}

	return schema, schema.load(db)
}

func (db *DBSchema) load(q queryer) error {
	if err := db.loadVariables(q); err != nil {
		return err
	}
	return db.loadTables(q)
}

func (db *DBSchema) loadVariables(q queryer) error {

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

func (db *DBSchema) loadTables(q queryer) error {

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
	Name string

	Pk *Index

	Columns []Column
	Indices []Index

	// add more stuff like keys
}

func (tbl *Table) load(q queryer) error {
	if err := tbl.loadColumns(q); err != nil {
		return err
	}

	sort.Sort(columnsByName(tbl.Columns))

	if err := tbl.loadIndices(q); err != nil {
		return err
	}

	sort.Sort(indexByKeyName(tbl.Indices))

	return nil
}

func (tbl *Table) loadColumns(q queryer) error {
	// sprintf'ing queries, because yolo (because prepared stmts dont
	// work for DDL)
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

func (tbl *Table) loadIndices(q queryer) error {
	// sprintf'ing queries, because yolo (because prepared stmts dont
	// work for DDL)
	rows, err := q.Query(fmt.Sprintf("show indexes in %s", tbl.Name))
	if err != nil {
		return fmt.Errorf("showing indices %q, %v", tbl.Name, err)
	}
	defer rows.Close()

	var parts []IndexPart
	for i := 0; rows.Next(); i++ {
		idx := IndexPart{}
		err := idx.scan(tbl, rows)
		if err != nil {
			return fmt.Errorf("scanning index %d, %v", i, err)
		}
		parts = append(parts, idx)
	}
	tbl.Pk, tbl.Indices = indicesFromParts(parts)

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

	fmt.Fprintf(w, "\tvar (\n")
	for _, idx := range tbl.Indices {
		fmt.Fprintf(w, "\t\t%s\n", idx.String())
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

/*
Indices!
*/

type IndexType int

const (
	startIndex IndexType = iota

	IndexBtree
	IndexFulltext
	IndexHash
	IndexRtree

	stopIndex
)

type Index struct {
	KeyName      string
	NonUnique    bool      // 0 if the index cannot contain duplicates, 1 if it can.
	Cardinality  int       // An estimate of the number of unique values in the index. This is updated by running ANALYZE TABLE or myisamchk -a. Cardinality is counted based on statistics stored as integers, so the value is not necessarily exact even for small tables. The higher the cardinality, the greater the chance that MySQL uses the index when doing joins.
	SubPart      *int      // The number of indexed characters if the column is only partly indexed, NULL if the entire column is indexed.
	Packed       *string   // Indicates how the key is packed. NULL if it is not.
	IndexType    IndexType // The index method used (BTREE, FULLTEXT, HASH, RTREE).
	Comment      string    // Information about the index not described in its own column, such as disabled if the index is disabled.
	IndexComment string    // Any comment provided for the index with a COMMENT attribute when the index was created.

	Columns []Column
	Parts   []IndexPart
}

func (idx Index) IsPrimary() bool { return idx.KeyName == "PRIMARY" }

func (idx Index) String() string {
	return fmt.Sprintf("%q (%v)", idx.KeyName, idx.Parts)
}

func indicesFromParts(parts []IndexPart) (pk *Index, indices []Index) {

	indexSet := make(map[string]Index)
	for _, part := range parts {
		idx, ok := indexSet[part.KeyName]
		if !ok {
			idx.KeyName = part.KeyName
			idx.NonUnique = part.NonUnique
			idx.Cardinality = part.Cardinality
			idx.SubPart = part.SubPart
			idx.Packed = part.Packed
			idx.IndexType = part.IndexType
			idx.Comment = part.Comment
			idx.IndexComment = part.IndexComment
		}
		idx.Parts = append(idx.Parts, part)
		idx.Columns = append(idx.Columns, part.Column)
		indexSet[idx.KeyName] = idx
	}
	var pkname string
	for _, idx := range indexSet {
		if idx.IsPrimary() {
			if pkname == "" {
				pkname = idx.KeyName
				continue
			} else {
				panic("multiple primary keys")
			}
		}
		indices = append(indices, idx)
	}
	if pkname != "" {
		idx := indexSet[pkname]
		pk = &idx
	}

	return
}

type IndexPart struct {
	Column Column

	Table        string    // The name of the table.
	NonUnique    bool      // 0 if the index cannot contain duplicates, 1 if it can.
	KeyName      string    // The name of the index. If the index is the primary key, the name is always PRIMARY.
	SeqInIndex   int       // The column sequence number in the index, starting with 1.
	ColumnName   string    // The column name.
	IsAscending  bool      // How the column is sorted in the index. In MySQL, this can have values “A” (Ascending) or NULL (Not sorted).
	Cardinality  int       // An estimate of the number of unique values in the index. This is updated by running ANALYZE TABLE or myisamchk -a. Cardinality is counted based on statistics stored as integers, so the value is not necessarily exact even for small tables. The higher the cardinality, the greater the chance that MySQL uses the index when doing joins.
	SubPart      *int      // The number of indexed characters if the column is only partly indexed, NULL if the entire column is indexed.
	Packed       *string   // Indicates how the key is packed. NULL if it is not.
	CanBeNull    bool      // Contains YES if the column may contain NULL values and '' if not.
	IndexType    IndexType // The index method used (BTREE, FULLTEXT, HASH, RTREE).
	Comment      string    // Information about the index not described in its own column, such as disabled if the index is disabled.
	IndexComment string    // Any comment provided for the index with a COMMENT attribute when the index was created.
}

func (idx *IndexPart) String() string {
	return idx.ColumnName
}

func (idx *IndexPart) scan(tbl *Table, rows *sql.Rows) error {
	var (
		nonUnique int
		collation string
		canBeNull string
		indexType string
	)

	err := rows.Scan(
		&idx.Table,
		&nonUnique,
		&idx.KeyName,
		&idx.SeqInIndex,
		&idx.ColumnName,
		&collation,
		&idx.Cardinality,
		&idx.SubPart,
		&idx.Packed,
		&canBeNull,
		&indexType,
		&idx.Comment,
		&idx.IndexComment,
	)
	idx.NonUnique = (nonUnique != 0)
	idx.IsAscending = (collation == "A")
	idx.CanBeNull = (canBeNull == "YES")
	switch indexType {
	case "BTREE":
		idx.IndexType = IndexBtree
	case "FULLTEXT":
		idx.IndexType = IndexFulltext
	case "HASH":
		idx.IndexType = IndexHash
	case "RTREE":
		idx.IndexType = IndexRtree
	}

	for _, col := range tbl.Columns {
		if col.Name == idx.ColumnName {
			idx.Column = col
			return err
		}
	}

	return fmt.Errorf("column %q doesn't exist for index %q", idx.ColumnName, idx.KeyName)
}

type columnsByName []Column

func (b columnsByName) Len() int      { return len(b) }
func (b columnsByName) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b columnsByName) Less(i, j int) bool {
	iname := b[i].Name
	jname := b[j].Name

	if iname == "id" {
		return true
	}
	if jname == "id" {
		return false
	}
	switch {
	case strings.Contains(iname, "_id") && !strings.Contains(jname, "_id"):
		return true
	case !strings.Contains(iname, "_id") && strings.Contains(jname, "_id"):
		return false
	}
	return iname < jname
}

type indexByKeyName []Index

func (b indexByKeyName) Len() int           { return len(b) }
func (b indexByKeyName) Less(i, j int) bool { return b[i].KeyName < b[j].KeyName }
func (b indexByKeyName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
