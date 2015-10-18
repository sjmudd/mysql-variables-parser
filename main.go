// script to parse the global variables page and generate an equivalent table structure
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sjmudd/mysql-variables-parser/parser"
)

var (
	flag_help    = flag.Bool("help", false, "Provide a usage message")
	flag_verbose = flag.Bool("verbose", false, "Make output verbose")
)

// very basic usage message
func usage(rc int) {
	fmt.Println(os.Args[0])
	fmt.Println("Script to parse the server-system-variables.html file and generate table defintions")
	fmt.Println("for the defined configuration settings")
	fmt.Println()
	fmt.Println("Usage: ", os.Args[0], "[--help] [--verbose] [<file_to_parse>] [<table_name>]")
	os.Exit(rc)
}

// main loop
func main() {
	var (
		filename  string
		tablename string
		parser    parser.Parser
		defaults  map[string]string
	)

	defaults = make(map[string]string)
	defaults["filename"] = "server-system-variables.html"
	defaults["tablename"] = "sysvars"

	flag.Parse()
	if *flag_help {
		usage(0)
	}
	if *flag_verbose {
		parser.SetVerbose()
	}

	args := flag.Args()
	switch len(args) {
	case 0:
		{
			filename = defaults["filename"]
			tablename = defaults["tablename"]
			if *flag_verbose {
				fmt.Println("no arguments provided")
				fmt.Println("- using default filename):", filename)
				fmt.Println("- using default tablename):", tablename)
			}
		}
	case 1:
		{
			filename = args[0]
			tablename = defaults["tablename"]
			if *flag_verbose {
				fmt.Println("1 argument (filename):", filename)
				fmt.Println("- using default tablename):", tablename)
			}
		}
	case 2:
		{
			filename = args[0]
			tablename = args[1]
			if *flag_verbose {
				fmt.Println("2 arguments provided")
				fmt.Println("- filename:", filename)
				fmt.Println("- tablename:", tablename)
			}
		}
	default:
		usage(1)
	}
	parser.Process(filename, tablename)
}
