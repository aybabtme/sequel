package generator

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/tabwriter"

	"github.com/aybabtme/sequel/reflector"
)

func createQuery(tbl reflector.Table) string {
	query := `
INSERT INTO %s (
%s
) VALUES ( %s )`

	cols := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(cols, 4, 8, 0, ' ', 0)
	for i, col := range setFields(tbl) {
		if i == 0 {
			fmt.Fprintf(w, "\t%s", escape(col.Name))
		} else {
			fmt.Fprintf(w, "\n\t, %s", escape(col.Name))
		}
	}
	w.Flush()

	vals := bytes.NewBuffer(nil)
	w = tabwriter.NewWriter(vals, 4, 8, 0, ' ', 0)
	for i := range setFields(tbl) {
		if i == 0 {
			fmt.Fprintf(w, "?")
		} else {
			fmt.Fprintf(w, ", ?")
		}
	}
	w.Flush()
	vals.WriteString("")

	return "`" + fmt.Sprintf(query, tbl.Name, cols.String(), vals.String()) + "`"
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
	for i, col := range tbl.Columns {
		if i == 0 {
			fmt.Fprintf(w, "\n\t%s", escape(col.Name))
		} else {
			fmt.Fprintf(w, "\n\t, %s", escape(col.Name))
		}
	}
	w.Flush()
	return selects.String()
}

func whereString(tbl reflector.Table) string {
	wheres := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(wheres, 4, 8, 0, ' ', 0)
	for i, col := range whereFields(tbl) {
		if i == 0 {
			fmt.Fprintf(w, "\n\t%s\t = ?", escape(col.Name))
		} else {
			fmt.Fprintf(w, "\n\tAND %s\t = ?", escape(col.Name))
		}
	}
	w.Flush()
	return wheres.String()
}

func whereIdxString(idx reflector.Index) string {
	wheres := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(wheres, 4, 8, 0, ' ', 0)
	for i, col := range idx.Columns {
		if i == 0 {
			fmt.Fprintf(w, "\n\t%s\t = ?", escape(col.Name))
		} else {
			fmt.Fprintf(w, "\n\tAND %s\t = ?", escape(col.Name))
		}
	}
	w.Flush()
	return wheres.String()
}

func setString(tbl reflector.Table) string {
	sets := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(sets, 4, 8, 0, ' ', 0)
	for i, col := range setFields(tbl) {
		if i == 0 {
			fmt.Fprintf(w, "\n\t%s\t = ?", escape(col.Name))
		} else {
			fmt.Fprintf(w, "\n\t, %s\t = ?", escape(col.Name))
		}
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
				return escape(col.ColumnName)
			}
		}
	}
	for _, idx := range tbl.Indices {
		for _, col := range idx.Parts {
			if col.IsAscending {
				return escape(col.ColumnName)
			}
		}
	}
	log.Printf("Table %q doesn't have an ordered column in one of its primary key or index. "+
		"Ordering will be done using column %q as a fallback.", tbl.Name, tbl.Columns[0].Name)
	return escape(tbl.Columns[0].Name)
}

func orderByIdxString(idx reflector.Index) string {
	// order by some index, or fallback to
	// the first column in the table
	for _, col := range idx.Parts {
		if col.IsAscending {
			return escape(col.ColumnName)
		}
	}
	log.Printf("Index %q doesn't have any ordered column."+
		"Ordering will be done using column %q as a fallback.",
		idx.KeyName, idx.Columns[0].Name)
	return escape(idx.Columns[0].Name)
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

func escape(colName string) string {
	switch strings.ToLower(colName) {
	default:
		return colName
	case "accessible",
		"add",
		"all",
		"alter",
		"analyze",
		"and",
		"asasc",
		"asensitive",
		"beforebetween",
		"bigint",
		"binaryblob",
		"both",
		"bycall",
		"cascade",
		"case",
		"changechar",
		"charactercheckcollate",
		"column",
		"conditionconstraint",
		"continue",
		"convert",
		"create",
		"crosscurrent_date",
		"current_time",
		"current_timestampcurrent_user",
		"cursor",
		"database",
		"databasesday_hour",
		"day_microsecond",
		"day_minute",
		"day_second",
		"dec",
		"decimal",
		"declare",
		"default",
		"delayed",
		"delete",
		"desc",
		"describe",
		"deterministic",
		"distinct",
		"distinctrow",
		"div",
		"double",
		"drop",
		"dual",
		"each",
		"else",
		"elseif",
		"enclosed",
		"escaped",
		"exists",
		"exit",
		"explain",
		"false",
		"fetchfloatfloat4",
		"float8",
		"for",
		"force",
		"foreign",
		"from",
		"fulltext",
		"get",
		"grantgroup",
		"having",
		"high_priorityhour_microsecond",
		"hour_minute",
		"hour_second",
		"if",
		"ignore",
		"in",
		"index",
		"infile",
		"innerinout",
		"insensitive",
		"insert",
		"int",
		"int1",
		"int2",
		"int3",
		"int4",
		"int8",
		"integer",
		"interval",
		"into",
		"io_after_gtids",
		"io_before_gtids",
		"is",
		"iterate",
		"join",
		"key",
		"keys",
		"kill",
		"leading",
		"leave",
		"left",
		"like",
		"limit",
		"linear",
		"linesload",
		"localtimelocaltimestamp",
		"lock",
		"long",
		"longblob",
		"longtext",
		"loop",
		"low_priority",
		"master_bind",
		"master_ssl_verify_server_certmatchmaxvalue",
		"mediumblob",
		"mediumintmediumtext",
		"middleintminute_microsecond",
		"minute_second",
		"mod",
		"modifies",
		"natural",
		"not",
		"no_write_to_binlog",
		"null",
		"numeric",
		"on",
		"optimize",
		"optimizer_costs",
		"option",
		"optionally",
		"or",
		"orderout",
		"outeroutfile",
		"partition",
		"precisionprimary",
		"procedure",
		"purgerangeread",
		"readsread_write",
		"real",
		"references",
		"regexp",
		"release",
		"rename",
		"repeat",
		"replace",
		"require",
		"resignal",
		"restrict",
		"return",
		"revoke",
		"right",
		"rlikeschema",
		"schemas",
		"second_microsecond",
		"select",
		"sensitive",
		"separatorset",
		"show",
		"signal",
		"smallint",
		"spatial",
		"specific",
		"sql",
		"sqlexception",
		"sqlstate",
		"sqlwarning",
		"sql_big_result",
		"sql_calc_found_rows",
		"sql_small_result",
		"ssl",
		"starting",
		"straight_jointable",
		"terminated",
		"then",
		"tinyblob",
		"tinyint",
		"tinytext",
		"to",
		"trailing",
		"trigger",
		"true",
		"undo",
		"unionunique",
		"unlock",
		"unsigned",
		"update",
		"usageuse",
		"using",
		"utc_date",
		"utc_time",
		"utc_timestamp",
		"values",
		"varbinaryvarchar",
		"varcharacter",
		"varying",
		"when",
		"wherewhilewith",
		"writexor",
		"year_month",
		"zerofill":
		return "`+\"`\"+`" + colName + "`+\"`\"+`"
	}
}
