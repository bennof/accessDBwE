// Copyright 2017 Benjamin 'Benno' Falkner. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package accessdbwe

import (
	"database/sql"
)

//
// Implementing new open function
//
func Open(driver, filen string) (*sql.DB, error) {
	var err error
	var db *sql.DB

	db, err = sql.Open(driver, filen)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}