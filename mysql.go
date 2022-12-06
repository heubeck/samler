/*
SaMLer - Smart Meter data colletor at the edge
Copyright (C) 2022  Florian Heubeck

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func InitializeMySQL(
	mysqlDSN string,
	tableName string,
) func(measurement Measurement) bool {
	fmt.Printf("Init MySQL")

	initialized := false
	var database *sql.DB

	runInit := func() bool {
		db, err := sql.Open("mysql", mysqlDSN)
		if err != nil {
			log.Printf("Failed to connect to MySQL: %s\n", err)
			return false
		}
		db.SetConnMaxLifetime(3 * time.Minute)
		db.SetMaxOpenConns(3)
		db.SetMaxIdleConns(1)

		if _, err := db.Exec("select now()"); err != nil {
			log.Printf("Could not connect to database: %s\n", err)
			return false
		}

		database = db
		return setupSchema(db, tableName)
	}

	sender := func(measurement Measurement) bool {
		if !initialized {
			initialized = runInit()
			if !initialized {
				return false
			}
		}

		debug("Sending to MySQL", &measurement)
		if _, err := database.Exec(fmt.Sprintf(`INSERT INTO %s
			(time, ident, value, unit, prefix, suffix)
			VALUES(?, ?, ?, ?, ?, ?)`, tableName),
			measurement.Time,
			measurement.Ident,
			measurement.Value,
			measurement.Unit,
			measurement.Prefix,
			measurement.Suffix,
		); err != nil {
			log.Printf("Failed sending to MySQL %s\n", err)
			initialized = false
			return false
		}
		return true
	}

	return sender
}

func setupSchema(db *sql.DB, tableName string) bool {
	log.Println("Creating database schema")
	schema := [...]string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id bigint NOT NULL AUTO_INCREMENT,
			time timestamp NOT NULL,
			ident varchar(10) NOT NULL,
			value double NOT NULL,
			unit varchar(5),
			prefix varchar(5),
			suffix varchar(5),
			PRIMARY KEY (id),
			INDEX idxTime (time DESC)
			)`, tableName),
	}

	for _, s := range schema {
		if _, err := db.Exec(s); err != nil {
			log.Printf("Failed to create schema: %s\n", err)
			return false
		}
	}
	return true
}
