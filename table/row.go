// manage rows of a table
package table

import (
	"fmt"
	"strings"
)

type Row struct {
	system_variable_name string
	cmd_line             string
	option_file          string
	system_var           string
	var_scope            string
	dynamic              string
	command_line_format  string
	default_value        string
	data_type            string
}

func (r *Row) SetSystemVariableName(name string) {
	r.system_variable_name = name
}
func (r *Row) SetCmdLine(name string) {
	r.cmd_line = name
}
func (r *Row) SetOptionFile(name string) {
	r.option_file = name
}
func (r *Row) SetSystemVar(name string) {
	r.system_var = name
}
func (r *Row) SetVarScope(name string) {
	r.var_scope = name
}
func (r *Row) SetDynamic(name string) {
	r.dynamic = name
}

func (r Row) Print() {
	fmt.Println("===")
	fmt.Println("system_variable_name:", r.system_variable_name)
	fmt.Println("cmd_line:            ", r.cmd_line)
	fmt.Println("option_file:         ", r.option_file)
	fmt.Println("system_var:          ", r.system_var)
	fmt.Println("var_scope:           ", r.var_scope)
	fmt.Println("dynamic:             ", r.dynamic)
	fmt.Println("command_line_format: ", r.command_line_format)
	fmt.Println("default_value:       ", r.default_value)
	fmt.Println("data_type:           ", r.data_type)
	fmt.Println("   ")
}

// IsEmpty returns true if all fields in the row are empty.
func (r Row) IsEmpty() bool {
	return len(r.system_variable_name)+
		len(r.cmd_line)+
		len(r.option_file)+
		len(r.system_var)+
		len(r.var_scope)+
		len(r.dynamic)+
		len(r.command_line_format)+
		len(r.default_value)+
		len(r.data_type) == 0
}

func (r Row) InsertStatement(table_name string) {
	column_names := []string{"system_variable_name", "cmd_line", "option_file", "system_var", "var_scope", "dynamic"}
	column_values := []string{r.system_variable_name, r.cmd_line, r.option_file, r.system_var, r.var_scope, r.dynamic}
	quoted_values := make([]string, 0, len(column_values))
	for i := range column_values {
		quoted_values = append(quoted_values, quote(column_values[i]))
	}

	s := "INSERT INTO " + table_name + " " +
		"(" + strings.Join(column_names, ",") + ")" +
		" VALUES " +
		"(" + strings.Join(quoted_values, ",") + ")" +
		";"

	fmt.Println(s)
}

// stupid quoting but good enough for me here.
func quote(s string) string {
	if s == "" {
		return "NULL"
	}
	return "'" + s + "'"
}

// return true if the two rows are the identical
func identical(r1, r2 Row) bool {
	return r1.system_variable_name == r2.system_variable_name &&
		r1.cmd_line == r2.cmd_line &&
		r1.option_file == r2.option_file &&
		r1.system_var == r2.system_var &&
		r1.var_scope == r2.var_scope &&
		r1.dynamic == r2.dynamic &&
		r1.command_line_format == r2.command_line_format &&
		r1.default_value == r2.default_value &&
		r1.data_type == r2.data_type
}

func showEmpty(s, comment string, answer bool) bool {
	//	fmt.Printf("empty('%s') len: %d '% x' => %v\n", s, len(s), s, answer)
	return answer
}

// a string is empty if it has no characters or only "spaces".
// We need to be careful here to recognise some special sequences
func empty(s string) bool {
	if len(s) == 0 {
		return showEmpty(s, "zero-length string", true)
	}
	if len(s) == 2 && s[0] == '\xc2' && s[1] == '\xa0' {
		return showEmpty(s, "c2 a0 combo empty string", true)
	}
	return showEmpty(s, "non-empty string", false)
}

func showDifferent(s1, s2 string, answer bool) bool {
	//	fmt.Println("different <",s1,"> <", s2, "> ->", answer)

	return answer
}

// count as different if they both have valid values which are not the identical but that neither is empty
func different(s1, s2 string) bool {
	if s1 == s2 {
		return showDifferent(s1, s2, false)
	}
	if empty(s1) || empty(s2) {
		return showDifferent(s1, s2, false)
	}
	return showDifferent(s1, s2, true)
}

// are the 2 rows mergeable? They are if no field is different
// note: different here means completely different values or one of the values is empty/blank
func mergeable(r1, r2 Row) bool {
	if different(r1.system_variable_name, r2.system_variable_name) ||
		different(r1.cmd_line, r2.cmd_line) ||
		different(r1.option_file, r2.option_file) ||
		different(r1.system_var, r2.system_var) ||
		different(r1.var_scope, r2.var_scope) ||
		different(r1.dynamic, r2.dynamic) ||
		different(r1.command_line_format, r2.command_line_format) ||
		different(r1.default_value, r2.default_value) ||
		different(r1.data_type, r2.data_type) {
		return false
	}
	return true
}

// merge 2 strings together taking into account whether one of them is empty
func merge(s1, s2 string) string {
	if empty(s1) {
		return s2
	}
	return s1
}

// Merge values together overwriting blank fields
func (r *Row) Merge(r2 Row) {
	r.cmd_line = merge(r.cmd_line, r2.cmd_line)
	r.option_file = merge(r.option_file, r2.option_file)
	r.system_var = merge(r.system_var, r2.system_var)
	r.var_scope = merge(r.var_scope, r2.var_scope)
	r.dynamic = merge(r.dynamic, r2.dynamic)
	r.command_line_format = merge(r.command_line_format, r2.command_line_format)
	r.default_value = merge(r.default_value, r2.default_value)
	r.data_type = merge(r.data_type, r2.data_type)
}
