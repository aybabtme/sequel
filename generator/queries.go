package generator

import (
	"bytes"
	"fmt"
	"github.com/aybabtme/sequel/reflector"
	"log"
	"text/tabwriter"
)

func createQuery(tbl reflector.Table) string {
	query := `
INSERT INTO %s VALUES ( %s )`

	vals := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(vals, 4, 8, 0, ' ', 0)
	for _, col := range setFields(tbl) {
		fmt.Fprintf(w, "\n\t%s\t = ?,", col.Name)
	}
	w.Flush()
	vals.WriteString("\n")

	return "`" + fmt.Sprintf(query, tbl.Name, vals.String()) + "`"
}

func retrieveQuery(tbl reflector.Table) string {

	query := `
SELECT %s
FROM %s
WHERE %s
LIMIT 1`

	return "`" + fmt.Sprintf(
		query,
		selectString(tbl),
		tbl.Name,
		whereString(tbl),
	) + "`"
}

func updateQuery(tbl reflector.Table) string {
	query := `
UPDATE %s
SET %s
WHERE %s`

	return "`" + fmt.Sprintf(
		query,
		tbl.Name,
		setString(tbl),
		whereString(tbl),
	) + "`"
}

func deleteQuery(tbl reflector.Table) string {
	query := `
DELETE FROM %s
WHERE %s`

	return "`" + fmt.Sprintf(
		query,
		tbl.Name,
		whereString(tbl),
	) + "`"
}

func listQuery(tbl reflector.Table) string {
	query := `
SELECT %s
FROM %s
ORDER BY %s
LIMIT 10000
OFFSET ?`

	return "`" + fmt.Sprintf(
		query,
		selectString(tbl),
		tbl.Name,
		orderByString(tbl),
	) + "`"
}

func listIndex(tbl reflector.Table, idx reflector.Index) string {
	query := `
SELECT %s
FROM %s
WHERE %s
ORDER BY %s
LIMIT 10000
OFFSET ?`

	return "`" + fmt.Sprintf(
		query,
		selectString(tbl),
		tbl.Name,
		whereIdxString(idx),
		orderByIdxString(idx),
	) + "`"
}

func selectString(tbl reflector.Table) string {
	selects := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(selects, 4, 8, 0, ' ', 0)
	for _, col := range tbl.Columns {
		fmt.Fprintf(w, "\n\t%s,", col.Name)
	}
	w.Flush()
	return selects.String()
}

func whereString(tbl reflector.Table) string {
	wheres := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(wheres, 4, 8, 0, ' ', 0)
	for _, col := range whereFields(tbl) {
		fmt.Fprintf(w, "\n\t%s\t = ?,", col.Name)
	}
	w.Flush()
	return wheres.String()
}

func whereIdxString(idx reflector.Index) string {
	wheres := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(wheres, 4, 8, 0, ' ', 0)
	for _, col := range idx.Columns {
		fmt.Fprintf(w, "\n\t%s\t = ?,", col.Name)
	}
	w.Flush()
	return wheres.String()
}

func setString(tbl reflector.Table) string {
	sets := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(sets, 4, 8, 0, ' ', 0)
	for _, col := range setFields(tbl) {
		fmt.Fprintf(w, "\n\t%s\t = ?,", col.Name)
	}
	w.Flush()
	return sets.String()
}

func orderByString(tbl reflector.Table) string {
	// order by pk index, or by some index, or fallback to
	// the first column in the table
	if tbl.Pk != nil {
		for _, col := range tbl.Pk.Parts {
			if col.IsAscending {
				return col.ColumnName
			}
		}
	}
	for _, idx := range tbl.Indices {
		for _, col := range idx.Parts {
			if col.IsAscending {
				return col.ColumnName
			}
		}
	}
	log.Printf("not ordering with an indexed key (%s)", tbl.Columns[0].Name)
	return tbl.Columns[0].Name
}

func orderByIdxString(idx reflector.Index) string {
	// order by some index, or fallback to
	// the first column in the table
	for _, col := range idx.Parts {
		if col.IsAscending {
			return col.ColumnName
		}
	}
	log.Printf("not ordering with an indexed key (%s)", idx.Columns[0].Name)
	return idx.Columns[0].Name
}

func whereFields(tbl reflector.Table) []reflector.Column {
	if tbl.Pk != nil {
		return tbl.Pk.Columns
	}
	return tbl.Columns
}

func setFields(tbl reflector.Table) []reflector.Column {
	var sets []reflector.Column
	for _, col := range tbl.Columns {
		if str, ok := col.Extra.([]byte); ok {
			if string(str) == "auto_increment" {
				continue
			}
		}
		sets = append(sets, col)
	}
	return sets
}
