package generator

//go:generate embed file -var clientTemplate --source db_client.gotmpl
const clientTemplate = "package {{.Name}}\n\nimport (\n    \"bytes\"\n    \"database/sql\"\n    \"database/sql/driver\"\n    \"encoding/json\"\n    \"fmt\"\n    \"log\"\n    \"time\"\n\n    \"github.com/go-sql-driver/mysql\"\n)\n\nvar (\n    _ Querier = &sql.DB{}\n    _ Querier = &sql.Tx{}\n)\n\n// Querier is the interface implemented by types that can\ntype Querier interface {\n    Exec(query string, args ...interface{}) (sql.Result, error)\n    Prepare(query string) (*sql.Stmt, error)\n    Query(query string, args ...interface{}) (*sql.Rows, error)\n    QueryRow(query string, args ...interface{}) *sql.Row\n}\n\nvar (\n    _ RowScanner = &sql.Row{}\n    _ RowScanner = &sql.Rows{}\n)\n\ntype RowScanner interface {\n    Scan(dest ...interface{}) error\n}\n\ntype Updater interface {\n    fields() []interface{}\n    cols() []string\n    FieldByColName(field string) (interface{}, error)\n}\n\n// Scan sets the columns named in cols into dst, using the data\n// in rs.\nfunc Scan(rs RowScanner, dst Updater, cols []string) error {\n    toScan := make([]interface{}, 0, len(cols))\n    for _, col := range cols {\n        field, err := dst.FieldByColName(col)\n        if err != nil {\n            return err\n        }\n        toScan = append(toScan, field)\n    }\n    return rs.Scan(toScan...)\n}\n\n\n// Null types\n\ntype NullInt64 sql.NullInt64\n\nvar (\n    _ json.Unmarshaler = &NullInt64{}\n    _ driver.Value     = &NullInt64{}\n)\n\nfunc NewInt64(i int64) NullInt64 {\n    return NullInt64{Int64: i, Valid: true}\n}\n\nfunc (n *NullInt64) Scan(value interface{}) error {\n    sqln := new(sql.NullInt64)\n    err := sqln.Scan(value)\n    *n = NullInt64(*sqln)\n    return err\n}\n\nfunc (n NullInt64) Value() (driver.Value, error) {\n    return sql.NullInt64(n).Value()\n}\n\nfunc (n *NullInt64) UnmarshalJSON(data []byte) error {\n    if bytes.Equal(data, []byte(\"null\")) {\n        n.Valid = false\n        return nil\n    }\n    err := json.Unmarshal(data, &n.Int64)\n    n.Valid = (err == nil)\n    return err\n}\n\nfunc (n NullInt64) MarshalJSON() ([]byte, error) {\n    if !n.Valid {\n        return []byte(\"null\"), nil\n    }\n    return json.Marshal(n.Int64)\n}\n\ntype NullString sql.NullString\n\nvar (\n    _ json.Unmarshaler = &NullString{}\n    _ driver.Value     = &NullString{}\n)\n\nfunc NewString(s string) NullString {\n    return NullString{String: s, Valid: true}\n}\n\nfunc (n *NullString) Scan(value interface{}) error {\n    sqln := new(sql.NullString)\n    err := sqln.Scan(value)\n    *n = NullString(*sqln)\n    return err\n}\n\nfunc (n NullString) Value() (driver.Value, error) {\n    return sql.NullString(n).Value()\n}\n\nfunc (n *NullString) UnmarshalJSON(data []byte) error {\n    if bytes.Equal(data, []byte(\"null\")) {\n        n.Valid = false\n        return nil\n    }\n    err := json.Unmarshal(data, &n.String)\n    n.Valid = (err == nil)\n    return err\n}\n\nfunc (n NullString) MarshalJSON() ([]byte, error) {\n    if !n.Valid {\n        return []byte(\"null\"), nil\n    }\n    return json.Marshal(n.String)\n}\n\ntype NullFloat64 sql.NullFloat64\n\nvar (\n    _ json.Unmarshaler = &NullFloat64{}\n    _ driver.Value     = &NullFloat64{}\n)\n\nfunc NewFloat64(f float64) NullFloat64 {\n    return NullFloat64{Float64: f, Valid: true}\n}\n\nfunc (n *NullFloat64) Scan(value interface{}) error {\n    sqln := new(sql.NullFloat64)\n    err := sqln.Scan(value)\n    *n = NullFloat64(*sqln)\n    return err\n}\n\nfunc (n NullFloat64) Value() (driver.Value, error) {\n    return sql.NullFloat64(n).Value()\n}\n\nfunc (n *NullFloat64) UnmarshalJSON(data []byte) error {\n    if bytes.Equal(data, []byte(\"null\")) {\n        n.Valid = false\n        return nil\n    }\n    err := json.Unmarshal(data, &n.Float64)\n    n.Valid = (err == nil)\n    return err\n}\n\nfunc (n NullFloat64) MarshalJSON() ([]byte, error) {\n    if !n.Valid {\n        return []byte(\"null\"), nil\n    }\n    return json.Marshal(n.Float64)\n}\n\ntype NullBool sql.NullBool\n\nvar (\n    _ json.Unmarshaler = &NullBool{}\n    _ driver.Value     = &NullBool{}\n)\n\nfunc NewBool(b bool) NullBool {\n    return NullBool{Bool: b, Valid: true}\n}\n\nfunc (n *NullBool) Scan(value interface{}) error {\n    sqln := new(sql.NullBool)\n    err := sqln.Scan(value)\n    *n = NullBool(*sqln)\n    return err\n}\n\nfunc (n NullBool) Value() (driver.Value, error) {\n    return sql.NullBool(n).Value()\n}\n\nfunc (n *NullBool) UnmarshalJSON(data []byte) error {\n    if bytes.Equal(data, []byte(\"null\")) {\n        n.Valid = false\n        return nil\n    }\n    err := json.Unmarshal(data, &n.Bool)\n    n.Valid = (err == nil)\n    return err\n}\n\nfunc (n NullBool) MarshalJSON() ([]byte, error) {\n    if !n.Valid {\n        return []byte(\"null\"), nil\n    }\n    return json.Marshal(n.Bool)\n}\n\ntype NullTime mysql.NullTime\n\nvar (\n    _ json.Unmarshaler = &NullTime{}\n    _ driver.Value     = &NullTime{}\n)\n\nfunc NewTime(t time.Time) NullTime {\n    return NullTime{Time: t, Valid: true}\n}\n\nfunc (n *NullTime) Scan(value interface{}) error {\n    sqln := new(mysql.NullTime)\n    err := sqln.Scan(value)\n    *n = NullTime(*sqln)\n    return err\n}\n\nfunc (n NullTime) Value() (driver.Value, error) {\n    return mysql.NullTime(n).Value()\n}\n\nfunc (n *NullTime) UnmarshalJSON(data []byte) error {\n    if bytes.Equal(data, []byte(\"null\")) {\n        n.Valid = false\n        return nil\n    }\n    err := json.Unmarshal(data, &n.Time)\n    n.Valid = (err == nil)\n    return err\n}\n\nfunc (n NullTime) MarshalJSON() ([]byte, error) {\n    if !n.Valid {\n        return []byte(\"null\"), nil\n    }\n    return json.Marshal(n.Time)\n}\n\n\n\nfunc isCommandOnTableDenied(err error) bool {\n    e, ok := err.(*mysql.MySQLError)\n    if !ok {\n        return false\n    }\n    return e.Number == 1142\n}\n\ntype Variables struct { {{range .Variables}}\n    {{.Name | camelize | export}} {{. | var_to_go_type}} {{end}}\n}\n\n{{$db := .Name | camelize | export}}\n\ntype {{$db}}DB struct {\n    Querier\n\n    Variables Variables\n{{range .Tables}}{{$table := .Name | camelize | pluralize | export}}\n    {{$table}} *{{$table}}{{end}}\n}\n\nfunc New{{$db}}DB(querier Querier) (*{{$db}}DB, error) {\n    var err error\n    db := &{{$db}}DB{\n        Querier: querier,\n        Variables: Variables{ {{range .Variables}}\n    {{.Name | camelize | export}}: {{. | var_to_go_value}}, {{end}}\n        },\n    }\n\n    {{range .Tables}}\n    {{$table := .Name | camelize | pluralize | export}}\n    db.{{$table}}, err = new{{$table}}(db)\n    if err != nil {\n        return nil, err\n    }\n    {{end}}\n\n    return db, nil\n}\n\n\n{{range .Tables}}\n{{$table := .Name | camelize | pluralize | export}}\n{{$tbl := .}}\n\nconst (\n    create{{$table}}SQL   = {{. | createQuery}}\n\n    {{if .Pk}}retrieve{{$table}}SQL = {{. | retrieveQuery }}\n    {{end}}\n    update{{$table}}SQL   = {{. | updateQuery }}\n\n    delete{{$table}}SQL   = {{. | deleteQuery }}\n\n    list{{$table}}SQL     = {{. | listQuery }}\n\n    {{range .Indices}}{{$idxname := .KeyName | camelize | export}}\n    list{{$table}}Idx{{$idxname}}SQL = {{listIndex $tbl .}}\n    {{end}}\n)\n\n// {{$table}} provides operations on {{.Name | camelize | pluralize}} stored in {{$db}}.\ntype {{$table}} struct {\n    db   Querier\n    Name string\n\n    create   *sql.Stmt\n    {{if .Pk}}retrieve *sql.Stmt {{end}}\n    update   *sql.Stmt\n    delete   *sql.Stmt\n    list     *sql.Stmt\n\n    {{range .Indices}}{{$idxname := .KeyName | camelize | export}}\n    idx{{.KeyName | camelize | export}} *sql.Stmt{{end}}\n}\n\nfunc new{{$table}}(db Querier) (*{{$table}}, error) {\n    var err error\n    tbl := &{{$table}}{db: db, Name: \"{{.Name}}\"}\n\n    bindings := []struct {\n        query string\n        stmt  **sql.Stmt\n    }{\n        {query: create{{$table}}SQL, stmt: &tbl.create},\n        {{if .Pk}}{query: retrieve{{$table}}SQL, stmt: &tbl.retrieve},{{end}}\n        {query: update{{$table}}SQL, stmt: &tbl.update},\n        {query: delete{{$table}}SQL, stmt: &tbl.delete},\n        {query: list{{$table}}SQL, stmt: &tbl.list},\n        // indices {{range .Indices}}{{$idxname := .KeyName | camelize | export}}\n        {query: list{{$table}}Idx{{$idxname}}SQL, stmt: &tbl.idx{{.KeyName | camelize | export}} }, {{end}}\n    }\n\n    for _, bind := range bindings {\n        (*bind.stmt), err = db.Prepare(bind.query)\n        switch {\n        case isCommandOnTableDenied(err):\n            log.Printf(\"unauthorized to perform query: %q\", bind.query)\n            return nil, nil // code trying to use this stmt should panic if they're not authorized\n        case err != nil:\n            return nil, fmt.Errorf(\"preparing query %q: %v\", bind.query, err)\n        }\n    }\n\n    return tbl, err\n}\n\n{{$datatype :=  $table | singularize }}\n\ntype {{$datatype}} struct { {{range .Columns}}\n    {{.Name | camelize | export}} {{. | col_to_go_type}} {{end}}\n}\n\nvar _ Updater = &{{$datatype}}{}\n\nfunc (d {{$datatype}}) cols() []string {\n    return []string{ {{range .Columns}}\n        \"{{.Name}}\",{{end}}\n    }\n}\n\nfunc (d {{$datatype}}) fields() []interface{}{\n    return []interface{}{ {{range .Columns}}\n        &d.{{.Name | camelize | export}}, {{end}}\n    }\n}\n\nfunc (d {{$datatype}}) FieldByColName(col string) (interface{}, error) {\n    switch col { {{range .Columns}}\n    case \"{{.Name}}\":\n        return &d.{{.Name | camelize | export}}, nil{{end}}\n    default:\n        return nil, fmt.Errorf(\"invalid column %q\", col)\n    }\n}\n\n\n{{if .Pk}}\n{{$pklen := len .Pk.Columns}}\n{{if eq $pklen 1}}\n{{$col := index .Pk.Columns 0}}\n{{$colname := $col.Name | camelize | export}}\n// Create a new {{$datatype}}.\nfunc (tbl *{{$table}}) Create(d *{{$datatype}}) error {\n\n    {{ $col := .Has \"updated_at\"}}\n    {{if $col}}\n    {{if $col.Nullable}}\n    // col is nullable {{$col.Nullable}}\n    d.CreatedAt = NewTime(time.Now())\n    {{else}}\n    d.CreatedAt = time.Now().UTC().Truncate(time.Second)\n    {{end}}\n    {{end}}\n\n    // skip the ID\n    res, err := tbl.create.Exec(d.fields()[1:]...)\n    if err != nil {\n        return err\n    }\n\n    id, err := res.LastInsertId()\n    if err != nil {\n        return err\n    }\n\n    d.{{$colname}} = int(id)\n    return nil\n}\n\n// Retrieve an existing {{$datatype}} by ID.\nfunc (tbl *{{$table}}) Retrieve(id int64) (*{{$datatype}}, bool, error) {\n\n    rs, err := tbl.retrieve.Query(id)\n    switch err {\n    default:\n        return nil, false, err\n    case sql.ErrNoRows:\n        return nil, false, nil\n    case nil:\n        defer rs.Close()\n    }\n    if !rs.Next() {\n        return nil, false, nil\n    }\n    d := &{{$datatype}}{}\n    return d, true, Scan(rs, d, d.cols())\n}\n\n// Update an existing {{$datatype}} by ID.\nfunc (tbl *{{$table}}) Update(d *{{$datatype}}) error {\n    {{ $col := .Has \"updated_at\"}}\n    {{if $col}}\n    {{if $col.Nullable}}\n    // col is nullable {{$col.Nullable}}\n    d.UpdatedAt = NewTime(time.Now())\n    {{else}}\n    d.UpdatedAt = time.Now().UTC().Truncate(time.Second)\n    {{end}}\n    {{end}}\n\n    _, err := tbl.update.Exec(append(d.fields()[1:], d.{{$colname}})...)\n    return err\n}\n\n// Delete an existing {{$datatype}} by ID.\nfunc (tbl *{{$table}}) Delete(d *{{$datatype}}) error {\n    _, err := tbl.delete.Exec(d.{{$colname}})\n    return err\n}\n{{else}}\n\n// Delete an existing {{$datatype}} by its fields (all fields must match)\nfunc (tbl *{{$table}}) Delete(d *{{$datatype}}) error {\n    _, err := tbl.delete.Exec(d.fields()...)\n    return err\n}\n{{end}}\n\n\n{{else}}\n\n// Create a new {{$datatype}}.\nfunc (tbl *{{$table}}) Create(d *{{$datatype}}) error {\n    {{if .Has \"created_at\"}}\n    d.CreatedAt = NewTime(time.Now())\n    {{end}}\n    _, err := tbl.create.Exec(d.fields()...)\n    return err\n}\n\n// Update an existing {{$datatype}} by its fields (all fields must match)\nfunc (tbl *{{$table}}) Update(d *{{$datatype}}) error {\n    {{if .Has \"updated_at\"}}\n    d.UpdatedAt = NewTime(time.Now())\n    {{end}}\n    _, err := tbl.update.Exec(d.fields()...)\n    return err\n}\n\n// Delete an existing {{$datatype}} by its fields (all fields must match)\nfunc (tbl *{{$table}}) Delete(d *{{$datatype}}) error {\n    _, err := tbl.delete.Exec(d.fields()...)\n    return err\n}\n\n{{end}}\n\n// List all {{$datatype}}s starting from an offset. Limited to 10k rows.\nfunc (tbl *{{$table}}) List(offset int) ([]{{$datatype}}, error) {\n    var list []{{$datatype}}\n\n    rows, err := tbl.list.Query(offset)\n    switch err {\n    default:\n        return nil, err\n    case sql.ErrNoRows:\n        return list, nil\n    case nil:\n        defer rows.Close()\n    }\n\n    for rows.Next() {\n        d := {{$datatype}}{}\n        if err := Scan(rows, &d, d.cols()); err != nil {\n            return list, err\n        }\n        list = append(list, d)\n    }\n\n    return list, rows.Err()\n}\n\n{{range .Indices}}{{$idxname := .KeyName | camelize | export}}\n// ListBy{{$idxname}} finds all {{$datatype}}s that match the query\n// pn the index `{{.KeyName}}`, starting at `offset`, limited to 10k rows.\nfunc (tbl *{{$table}}) ListBy{{$idxname}}({{. | idx_list_args}}, offset int) ([]{{$datatype}}, error) {\n    var list []{{$datatype}}\n\n    rows, err := tbl.idx{{$idxname}}.Query({{. | idx_query_args}}, offset)\n    switch err {\n    default:\n        return nil, err\n    case sql.ErrNoRows:\n        return list, nil\n    case nil:\n        defer rows.Close()\n    }\n\n    for rows.Next() {\n        d := {{$datatype}}{}\n        if err := Scan(rows, &d, d.cols()); err != nil {\n            return list, err\n        }\n        list = append(list, d)\n    }\n\n    return list, rows.Err()\n}\n{{end}}\n\n{{end}}\n"
