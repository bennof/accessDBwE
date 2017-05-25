// Copyright 2017 Benjamin 'Benno' Falkner. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"bufio"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-adodb"
	"github.com/bennof/AccessDBwE"
	"os"
	"strings"
)


func main () {
	var dbstring string
	var cmd bytes.Buffer
	var exit bool
	
	fmt.Println("SQL Shell")

	var i int = 1
	for i < len(os.Args) {
		//fmt.Println(i,os.Args[i])
		switch os.Args[i] {
			case "-h" :
				fmt.Println("HELP:")
				return 
			default:
				dbstring = os.Args[i]
		}
		i++
	}

	// open db file
	if dbstring == "" {fmt.Println("No DB to open"); os.Exit(1)}
	fmt.Println("Open DB:",dbstring)
	db, err := accessdbwe.Open("adodb","Provider=Microsoft.ACE.OLEDB.12.0;Data Source="+dbstring)
	//db, err := sql.Open("adodb","Provider=Microsoft.ACE.OLEDB.12.0;Data Source="+dbstring)
	if err != nil { fmt.Println(err); os.Exit(1)}

	reader := bufio.NewReader(os.Stdin)
	w := csv.NewWriter(os.Stdout)
	for !exit {
		os.Stdout.WriteString("sql>")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		text = strings.Trim(text," \t\n\r")
		// fmt.Println(">>>> ",text) // debug

		switch {
			case text == "":  continue
			case strings.HasPrefix(text,"exit"): exit=true
			case text[len(text)-1] == ';':
				cmd.WriteString(" ")
				cmd.WriteString(text)
				rows, err := db.Query(cmd.String())
				if err != nil {
					fmt.Println("     ERROR:",err)
					fmt.Println("     ",cmd.String())
				} else {
					fmt.Println("     SUCCESS")
				}

				if  err == nil && rows != nil {
					names, err := rows.Columns()
					if err != nil { 
						fmt.Println("     ERROR:",err) 
					} else {
						w.Write(names)
						data := make([]interface{},len(names))
						for rows.Next() {
							rows.Scan(data...)
							w.Write(names)
						}
					}	
				}

				cmd.Reset()
			default: 
				fmt.Println("add cmd");
				cmd.WriteString(" ")
				cmd.WriteString(text)
		}
	}
	db.Close()
}

