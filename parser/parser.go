// package parser contains information for parsing the tokens in the input file
package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"code.google.com/p/go.net/html"

	"github.com/sjmudd/mysql-variables-parser/sysvar"
	"github.com/sjmudd/mysql-variables-parser/table"
)

const (
	defaultTableName = "server_system_variables"
	// TokenHistorySize represents the size of the token history we remember
	TokenHistorySize = 15
)

/* sample reformatted output from http://www.freeformatter.com/html-formatter.html makes things easier to understand.

   <table summary="Options for flush" border="1">    <<======= name of variable
     <colgroup>
       <col class="title">
       <col class="vt">
       <col class="vd">
       <col class="v">
     </colgroup>
     <tbody>
       <tr>
         <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td>    <<=== command line
         <td colspan="3"><code class="literal">--flush</code></td>                              <<=== command line
       </tr>
       <tr>
         <td scope="row" rowspan="3"><span class="bold"><strong>System Variable</strong></span></td>
         <td><span class="bold"><strong>Name</strong></span></td>
         <td colspan="2"><code class="literal"><a class="link" href="server-system-variables.html#sysvar_flush">flush</a></code></td>
       </tr>
       <tr>
         <td scope="row"><span class="bold"><strong>Variable Scope</strong></span></td>      <<==== scope
         <td colspan="2">Global</td>                                                         <<==== scope
       </tr>
       <tr>
         <td scope="row"><span class="bold"><strong>Dynamic Variable</strong></span></td>    <<=== dynamic
         <td colspan="2">Yes</td>                                                            <<=== dynamic
       </tr>
       <tr>
         <td scope="row" rowspan="2"><span class="bold"><strong>Permitted Values</strong></span></td>
         <td><span class="bold"><strong>Type</strong></span></td>                            <<=== type
         <td colspan="2"><code class="literal">boolean</code></td>                           <<=== type
       </tr>
       <tr>
         <td scope="row"><span class="bold"><strong>Default</strong></span></td>             <<==== default
         <td colspan="2"><code class="literal">OFF</code></td>                               <<==== default
       </tr>
     </tbody>
   </table>

*/

// Handler is a function which processes a token
type Handler func(html.Token) error

// TokenHistory is a slice of tokens
type TokenHistory []html.Token

// Parser contains information about the context we are processing
type Parser struct {
	tokenizer    *html.Tokenizer
	tokenCount   int
	handler      Handler
	tokenHistory TokenHistory
	table        *table.Table
	row          table.Row
	rowNum       int
	colNum       int
	sysvarInfo   sysvar.Info
	verbose      bool
}

// Store the last TokenHistorySize tokens in tokenHistory so we can look back
// position 0 is the current token, 1 is the previous one etc..
func (c *Parser) getToken() html.Token {
	token := c.tokenizer.Token()

	th := make(TokenHistory, 0, TokenHistorySize)
	th = append(th, token)

	if c.tokenHistory != nil {
		for i := range c.tokenHistory {
			if len(c.tokenHistory) >= TokenHistorySize {
				break
			}
			th = append(th, c.tokenHistory[i])
		}
	}
	c.tokenHistory = th
	c.tokenCount++
	if c.verbose {
		c.printTokenHistory()
	}

	return token
}

func (c *Parser) printTokenHistory() {
	fmt.Println("tokenHistory length:", len(c.tokenHistory))
	for i := range c.tokenHistory {
		fmt.Println(" ", i, c.tokenHistory[i].Type, c.tokenHistory[i])
	}
	fmt.Println("tokenHistory: END")
}

// Process parses the file consuming tokens and finally returning the SQL statements to build a table
func (c *Parser) Process(filename string, tablename string) {
	var err error
	var fi *os.File

	c.table = table.NewTable(tablename)

	if filename == "-" {
		fi = os.Stdin
	} else {
		fi, err = os.Open(filename)
		if err != nil {
			panic(err)
		}
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	r := bufio.NewReader(fi) // make a read buffer

	c.tokenizer = html.NewTokenizer(r)
	c.handler = c.WaitingForTable

	err = nil
	done := false
	for !done {
		_ = c.tokenizer.Next() // token type - do we need to do this in 2 calls and throw away the token type ?
		token := c.getToken()

		if c.verbose {
			fmt.Println("Process(): tokenCount:", c.tokenCount, ", handler:", c.handler, ", err:", err)
		}
		err = c.handler(token)
		if c.handler == nil || err != nil {
			done = true
		}
	}

	if c.verbose {
		fmt.Println("Process completed after consuming", c.tokenCount, "tokens")
	}
	if err != nil {
		log.Panic("Failed to consume tokens:", err)
	}
}

// WaitingForTable processes tokens waiting for the main table to start
func (c *Parser) WaitingForTable(token html.Token) error {
	if c.verbose {
		fmt.Println("WaitingForTable(", token, ")")
	}

	if token.Data == "table" &&
		len(token.Attr) > 0 &&
		token.Attr[0].Key == "summary" &&
		token.Attr[0].Val == "System Variable Summary" {
		c.handler = c.ProcessingTable
		if c.verbose {
			fmt.Println("STATE CHANGE: WaitingForTable() - change handler to: ProcessingTable")
		}
	}
	return nil
}

// ProcessingTable process the content of the main table
func (c *Parser) ProcessingTable(token html.Token) error {
	switch token.Type {
	case html.StartTagToken:
		{
			switch token.Data {
			case "tr":
				c.NewRow()
			case "td":
				c.NewCol()
			}
		}
	case html.EndTagToken:
		{
			switch token.Data {
			case "table":
				{
					c.handler = c.WaitingForDetails
					if c.verbose {
						fmt.Println("STATE CHANGE: ProcessingTable() - change handler to: WaitingForDetails")
					}
					c.ResetRowCounters()
				}
			case "tr":
				c.SaveRow()
				//      printBit(token)
			}
		}
	case html.TextToken:
		c.SetText(token)
	case html.CommentToken:
		//		utils.PrintToken(token)
	case html.DoctypeToken:
		//		utils.PrintToken(token)
	case html.SelfClosingTagToken:
	default: /* do nothing */
	}

	return nil
}

// WaitingForDetails processes the token while waiting for details
func (c *Parser) WaitingForDetails(token html.Token) error {
	switch token.Type {
	case html.StartTagToken:
		{
			switch token.Data {
			// <table summary="Options for version_compile_os" border="1">
			case "table":
				{
					sysvarName, found := returnSysvarName(token)
					if found {
						if c.verbose {
							fmt.Println("-- sysvar name:", sysvarName)
						}
						c.sysvarInfo.SaveName(sysvarName)
					}
				}
			default:
				/* do nothing */
			}
		}
	case html.EndTagToken:
		{
			switch token.Data {
			case "html":
				{
					c.handler = nil
					if c.verbose {
						fmt.Println("STATE CHANGE: found final </html>, so finish processing file")
					}
					c.table.MysqlDump()
					//					// dump from data
					//					fmt.Println("-- XXXXXXXX --")
					//					c.table.MysqlDumpFromSysvars(c.sysvarInfo.ColumnTypes(), c.sysvarInfo.CmdLines(), c.sysvarInfo.Scopes(), c.sysvarInfo.Defaults(), c.sysvarInfo.Dynamics())
					c.ResetRowCounters()
				}
			case "tr":
				{
					columnType, found := returnSysvarType(c.tokenHistory)
					if found {
						if c.verbose {
							fmt.Println("--        type:", columnType)
						}
						c.sysvarInfo.SaveType(columnType)
						return nil
					}
					cmdLine, found := returnCommandLine(c.tokenHistory)
					if found {
						if c.verbose {
							fmt.Println("-- sysvar type:", cmdLine)
						}
						c.sysvarInfo.SaveCommandLine(cmdLine)
						return nil
					}
					scope, found := returnSysvarScope(c.tokenHistory)
					if found {
						if c.verbose {
							fmt.Println("--       scope:", scope)
						}
						c.sysvarInfo.SaveScope(scope)
						return nil
					}
					defaultVal, found := returnSysvarDefault(c.tokenHistory)
					if found {
						if c.verbose {
							fmt.Println("--     default:", defaultVal)
						}
						c.sysvarInfo.SaveDefault(defaultVal)
						return nil
					}
					dynamic, found := returnSysvarDynamic(c.tokenHistory)
					if found {
						if c.verbose {
							fmt.Println("--     dynamic:", dynamic)
						}
						c.sysvarInfo.SaveDynamic(dynamic)
						return nil
					}
				}
			default: /* do nothing */
			}
		}
	default: /* do nothing */
	}

	return nil
}

// This returns the sysvar name.
// <table summary="Options for flush" border="1">
//                 0123456789012345678
//                           1
func returnSysvarName(token html.Token) (string, bool) {
	if token.Type == html.StartTagToken &&
		token.Data == "table" &&
		len(token.Attr) > 0 &&
		token.Attr[0].Key == "summary" &&
		len(token.Attr[0].Val) > 12 &&
		token.Attr[0].Val[0:11] == "Options for" {
		return token.Attr[0].Val[12:], true
	}
	return "", false
}

//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--default_tmp_storage_engine=name</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--default_week_format=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--delay-key-write[=name]</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--delayed_insert_limit=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--delayed_insert_timeout=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--delayed_queue_size=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--disconnect_on_expired_password=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--div_precision_increment=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--engine-condition-pushdown</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--event-scheduler[=value]</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--expire_logs_days=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--explicit_defaults_for_timestamp=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--flush</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--flush_time=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--ft_boolean_syntax=name</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--ft_max_word_len=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--ft_min_word_len=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--ft_query_expansion_limit=#</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--ft_stopword_file=file_name</code></td></tr>
//     <td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--general-log</code></td></tr>
// <tr><td scope="row"><span class="bold"><strong> Command-Line Format</strong></span> </td> <td colspan="3"><code class="literal">--net_write_timeout=#</code></td> </tr>
// <tr><td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--big-tables</code></td></tr>
// <tr><td scope="row"><span class="bold"><strong>Command-Line Format</strong></span></td><td colspan="3"><code class="literal">--flush</code></td></tr>
//  13         12            11               10            9            8        7    6           5                  4             3     2     1    0
//
func returnCommandLine(th TokenHistory) (string, bool) {
	if th != nil &&
		len(th) >= 14 &&
		th[0].Type == html.EndTagToken &&
		th[0].Data == "tr" &&
		th[1].Type == html.EndTagToken &&
		th[1].Data == "td" &&
		th[2].Type == html.EndTagToken &&
		th[2].Data == "code" &&
		th[3].Type == html.TextToken && // th[3].Data <<--- is what I'm looking for

		th[4].Type == html.StartTagToken && th[4].Data == "code" &&
		th[5].Type == html.StartTagToken && th[5].Data == "td" &&
		th[6].Type == html.EndTagToken && th[6].Data == "td" &&

		th[9].Type == html.TextToken &&
		(th[9].Data == "Command-Line Format" || th[9].Data == " Command-Line Format") && // horrible hack !

		th[10].Type == html.StartTagToken &&
		th[10].Data == "strong" &&
		th[11].Type == html.StartTagToken &&
		th[11].Data == "span" &&
		th[12].Type == html.StartTagToken &&
		th[12].Data == "td" &&
		th[13].Type == html.StartTagToken &&
		th[13].Data == "tr" {

		// fmt.Println("command line:", th[3].Data)
		return th[3].Data, true
	}
	return "", false
}

// this is how we recognise a type:
// <td><span class="bold"><strong>Type</strong></span></td><td colspan="2"><code class="literal">integer</code></td></tr>
//                                  9      8       7    6         5                    4            3       2    1    0
// <td><span class="bold"><strong>Type</strong></span></td><td colspan="2"><code class="literal">integer</code></td></tr>
func returnSysvarType(th TokenHistory) (string, bool) {
	if th != nil &&
		len(th) >= 10 && // 10
		th[0].Type == html.EndTagToken &&
		th[0].Data == "tr" &&
		th[1].Type == html.EndTagToken &&
		th[1].Data == "td" &&
		th[2].Type == html.EndTagToken &&
		th[2].Data == "code" &&
		th[3].Type == html.TextToken && // th[3].Data <<--- is what I'm looking for
		th[4].Type == html.StartTagToken &&
		th[4].Data == "code" &&
		th[5].Type == html.StartTagToken &&
		th[5].Data == "td" &&
		th[6].Type == html.EndTagToken &&
		th[6].Data == "td" &&
		th[7].Type == html.EndTagToken &&
		th[7].Data == "span" &&
		th[8].Type == html.EndTagToken &&
		th[8].Data == "strong" &&
		th[9].Type == html.TextToken &&
		th[9].Data == "Type" {
		return th[3].Data, true
	}
	return "", false
}

// <tr><td scope="row"><span class="bold"><strong>Variable Scope</strong></span></td><td colspan="2">Global</td></tr>
//  11         10          9                   8         7           6      5      4   3                2     1   0
func returnSysvarScope(th TokenHistory) (string, bool) {
	if th != nil &&
		len(th) >= 12 &&
		th[0].Type == html.EndTagToken && th[0].Data == "tr" &&
		th[1].Type == html.EndTagToken && th[1].Data == "td" &&
		th[2].Type == html.TextToken && // th[2].Data <<-- is what I'm looking for
		th[3].Type == html.StartTagToken && th[3].Data == "td" &&
		th[4].Type == html.EndTagToken && th[4].Data == "td" &&
		th[5].Type == html.EndTagToken && th[5].Data == "span" &&
		th[6].Type == html.EndTagToken && th[6].Data == "strong" &&
		th[7].Type == html.TextToken && th[7].Data == "Variable Scope" &&
		th[8].Type == html.StartTagToken && th[8].Data == "strong" &&
		th[9].Type == html.StartTagToken && th[9].Data == "span" &&
		th[10].Type == html.StartTagToken && th[10].Data == "td" &&
		th[11].Type == html.StartTagToken && th[11].Data == "tr" {
		return th[2].Data, true
	}
	return "", false
}

// <tr><td scope="row"><span class="bold"><strong>Default</strong></span></td><td colspan="2"><code class="literal">28800</code></td></tr>
//  13       12              11             10      9       8       7     6       5                    4              3     2     1   0
func returnSysvarDefault(th TokenHistory) (string, bool) {
	if th != nil &&
		len(th) >= 14 &&
		th[0].Type == html.EndTagToken && th[0].Data == "tr" &&
		th[1].Type == html.EndTagToken && th[1].Data == "td" &&
		th[2].Type == html.EndTagToken && th[2].Data == "code" &&
		th[3].Type == html.TextToken && // th[3].Data <<-- is what I'm looking for
		th[4].Type == html.StartTagToken && th[4].Data == "code" &&
		th[5].Type == html.StartTagToken && th[5].Data == "td" &&
		th[6].Type == html.EndTagToken && th[6].Data == "td" &&
		th[7].Type == html.EndTagToken && th[7].Data == "span" &&
		th[8].Type == html.EndTagToken && th[8].Data == "strong" &&
		th[9].Type == html.TextToken && th[9].Data == "Default" &&
		th[10].Type == html.StartTagToken && th[10].Data == "strong" &&
		th[11].Type == html.StartTagToken && th[11].Data == "span" &&
		th[12].Type == html.StartTagToken && th[12].Data == "td" &&
		th[13].Type == html.StartTagToken && th[13].Data == "tr" {
		return th[3].Data, true
	}
	return "", false
}

//  <tr><td scope="row"><span class="bold"><strong>Dynamic Variable</strong></span></td><td colspan="2">Yes</td></tr>
//   11         10              9              8            7          6       5     4          3        2   1    0
func returnSysvarDynamic(th TokenHistory) (string, bool) {
	if th != nil &&
		len(th) >= 12 &&
		th[0].Type == html.EndTagToken && th[0].Data == "tr" &&
		th[1].Type == html.EndTagToken && th[1].Data == "td" &&
		th[2].Type == html.TextToken && // th[2].Data <<-- is what I'm looking for
		th[3].Type == html.StartTagToken && th[3].Data == "td" &&
		th[4].Type == html.EndTagToken && th[4].Data == "td" &&
		th[5].Type == html.EndTagToken && th[5].Data == "span" &&
		th[6].Type == html.EndTagToken && th[6].Data == "strong" &&
		th[7].Type == html.TextToken && th[7].Data == "Dynamic Variable" &&
		th[8].Type == html.StartTagToken && th[8].Data == "strong" &&
		th[9].Type == html.StartTagToken && th[9].Data == "span" &&
		th[10].Type == html.StartTagToken && th[10].Data == "td" &&
		th[11].Type == html.StartTagToken && th[11].Data == "tr" {
		return th[2].Data, true
	}
	return "", false
}

// NewRow increments the row count and sets the column number to 0
func (c *Parser) NewRow() {
	c.rowNum++
	c.colNum = 0
	// fmt.Println("NEW_ROW: Row", c.rowNum)
}

// NewCol increments the column count
func (c *Parser) NewCol() {
	c.colNum++
	//	fmt.Println("NEW_COL: Row:", s.rowNum, "Col:", s.colNum)
}

// RowNo returns the current row number
func (c Parser) RowNo() int {
	return c.rowNum
}

// ColNo returns the current column number
func (c Parser) ColNo() int {
	return c.colNum
}

// SetText puts the text in the appropriate field
func (c *Parser) SetText(token html.Token) {
	//	utils.PrintToken(token)

	switch c.colNum {
	case 1:
		c.row.SetSystemVariableName(token.Data)
	case 2:
		c.row.SetCmdLine(token.Data)
	case 3:
		c.row.SetOptionFile(token.Data)
	case 4:
		c.row.SetSystemVar(token.Data)
	case 5:
		c.row.SetVarScope(token.Data)
	case 6:
		c.row.SetDynamic(token.Data)
	default:
		// ignore failure for now
	}
	//	s.row.Print()
}

// PrintRow prints the row if we have some data
func (c Parser) PrintRow() {
	if !c.row.IsEmpty() {
		c.row.Print()
	}
}

// SaveRow saves the row details in the parser.
func (c *Parser) SaveRow() {
	//	fmt.Println()
	if c.colNum == 6 {
		// s.PrintRow()
		c.table.AppendRow(c.row)
		c.row = table.Row{}
		// fmt.Printf("Saved row to %s, rows: %d\n", s.table.name, s.table.Rows())
	} else {
		c.rowNum-- // hack but should stop us increasing row numbers
		// fmt.Println("Ignoring row with", s.colNum, "columns")
	}
}

// ResetRowCounters resets the row counters
func (c *Parser) ResetRowCounters() {
	c.colNum = 0
	c.rowNum = 0
}

// SetVerbose makes logging more verbose
func (c *Parser) SetVerbose() {
	c.verbose = true
}
