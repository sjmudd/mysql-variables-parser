package table

import (
	"fmt"
	"sort"

	"github.com/sjmudd/mysql-variables-parser/sysvar"
)

type Table struct {
	name         string
	rows         []Row
	varNameToRow map[string]int // maps the variable name to the row it's stored in.
}

// create a new table with the given name
func NewTable(name string) *Table {
	fmt.Println("-- New table:" + name)
	t := new(Table)
	t.name = name
	t.varNameToRow = make(map[string]int)
	return t
}

// return the number of rows in the table
func (t Table) Rows() int {
	return len(t.rows)
}

// generate a create table statement, currently hard-coded
func (t Table) CreateTableStatement() {
	s := `-- Create table entry
DROP TABLE IF EXISTS %s;
CREATE TABLE %s (
    system_variable_name varchar(255) NOT NULL,
    cmd_line varchar(255) DEFAULT NULL,
    option_file varchar(50) DEFAULT NULL,
    system_var varchar(50) DEFAULT NULL,
    var_scope varchar(50) DEFAULT NULL,
    dynamic varchar(50) DEFAULT NULL,
    data_type varchar(50) DEFAULT NULL,
    PRIMARY KEY (system_variable_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
	fmt.Printf(s, t.name, t.name)
}

// create the INSERT statements for the rows in the table
func (t Table) InsertStatements() {
	if len(t.rows) > 0 {
		fmt.Println("-- Insert rows")
	}
	for i := range t.rows {
		if !t.rows[i].IsEmpty() {
			t.rows[i].InsertStatement(t.name)
		}
	}
}

// AppendRow Appends a row to the table if the variable name has not been seen.
// If it has then it will check if the values are the identical and do nothing.
func (t *Table) AppendRow(row Row) {
	if t.rows == nil {
		t.rows = make([]Row, 0, 100)
	}

	if i, ok := t.varNameToRow[row.system_variable_name]; ok {
		if !identical(t.rows[i], row) {
			// fmt.Println("Duplicate row:", row.system_variable_name)
			if mergeable(t.rows[i], row) {
				// fmt.Println("Merging:  ", t.rows[i])
				// fmt.Println("With:     ", row)
				row.Merge(t.rows[i])
				t.rows[i] = row
				// fmt.Println("Gives:    ", row)
			} else {
				fmt.Println("NOT THE SAME AND NOT MERGEABLE")
				fmt.Println("previous:", t.rows[i])
				fmt.Println("latest:  ", row)
			}
		} else {
			// fmt.Println("Duplicate row:", row.system_variable_name, "matches: ignoring")
		}
	} else {
		t.varNameToRow[row.system_variable_name] = len(t.rows)
		t.rows = append(t.rows, row)
	}
}

// Print the contents of the rows in the table.
func (t Table) Print() {
	for i := range t.rows {
		t.rows[i].Print()
	}
}

// Generate the equivalent of a mysqldump <db> <table>.
func (t Table) MysqlDump() {
	t.CreateTableStatement()
	t.InsertStatements()
}

type Keys []string

func (k Keys) Len() int { return len(k) }
func (k Keys) Less(i, j int) bool {
	return (k[i] < k[j])
}
func (k Keys) Swap(i, j int) { k[i], k[j] = k[j], k[i] }

// do a mysql dump from the collected sysvar info
func (t Table) MysqlDumpFromSysvars(types, cmdlines, scopes, defaults, dynamics sysvar.Types) {
	m := make(sysvar.Types)

	t.CreateTableStatement()

	// combine all the variable names we have together
	for k, v := range types {
		m[k] = v
	}
	for k, v := range cmdlines {
		m[k] = v
	}
	for k, v := range scopes {
		m[k] = v
	}
	for k, v := range defaults {
		m[k] = v
	}
	for k, v := range dynamics {
		m[k] = v
	}

	// sort keys
	fmt.Println("-- PENDING SORTING KEYS")
	fmt.Println("-- start dump from sysvars")
	var keys = make(Keys, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	for i := range keys {
		var r Row
		r.SetSystemVariableName(keys[i])

		if v, found := scopes[keys[i]]; found {
			r.SetVarScope(v)
		}
		if v, found := cmdlines[keys[i]]; found {
			r.SetCmdLine(v)
		}
		if v, found := dynamics[keys[i]]; found {
			r.SetDynamic(v)
		}
		if !r.IsEmpty() {
			r.InsertStatement(t.name)
		}
	}

	fmt.Println("-- end dump from sysvars")
}
