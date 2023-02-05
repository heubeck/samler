/*
SaMLer - Smart Meter data colletor at the edge
Copyright (C) 2023  Florian Heubeck

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
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFailingMySqlSend(t *testing.T) {
	// Given
	sender := InitializeMySQL("", "")
	m := Measurement{}

	// When
	success := sender(m)

	// Then
	if success != false {
		t.Fatal()
	}
}

func TestSuccessfulMySQLSend(t *testing.T) {
	// Setup testcontainer
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8",
		ExposedPorts: []string{"3306/tcp"},
		WaitingFor:   wait.ForExposedPort(),
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "samler",
		},
	}
	ctx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "3306/tcp")

	// Given
	dsn := fmt.Sprintf("root:password@tcp(%s:%d)/samler", host, port.Int())

	sender := selectBackend(map[string]string{
		Backend:    "mysql",
		MySqlDSN:   dsn,
		MySqlTable: "measures",
	})

	m := Measurement{
		Time:   time.Now(),
		Value:  22.5,
		Unit:   "°",
		Ident:  "0.8.15",
		Prefix: "1-2",
		Suffix: "667",
	}

	// When
	success := sender(m)

	// Then
	if !success {
		t.Fatal()
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}

	res, err := db.Query("select ident, value, unit, prefix, suffix from measures")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Next() {
		t.Fatal("Didn't get a result")
	}

	var value float32
	var ident, unit, prefix, suffix string
	if err := res.Scan(&ident, &value, &unit, &prefix, &suffix); err != nil {
		t.Fatal(err)
	}

	if value != 22.5 || unit != "°" || prefix != "1-2" || suffix != "667" || ident != "0.8.15" {
		t.Fatal()
	}
}
