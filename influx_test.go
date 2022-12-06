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
	"context"
	"fmt"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFailingInfluxSend(t *testing.T) {
	// Given
	sender := InitializeInflux("", "", "", "", "")
	m := Measurement{}

	// When
	success := sender(m)

	// Then
	if success != false {
		t.Fatal()
	}
}

func TestSuccessfulInfluxSend(t *testing.T) {
	// Setup testcontainer
	token := "this-is-secret"
	req := testcontainers.ContainerRequest{
		Image:        "influxdb:2.5",
		ExposedPorts: []string{"8086/tcp"},
		WaitingFor:   wait.ForExposedPort(),
		Env: map[string]string{
			"DOCKER_INFLUXDB_INIT_MODE":        "setup",
			"DOCKER_INFLUXDB_INIT_USERNAME":    "root",
			"DOCKER_INFLUXDB_INIT_PASSWORD":    "password",
			"DOCKER_INFLUXDB_INIT_ORG":         "samler",
			"DOCKER_INFLUXDB_INIT_BUCKET":      "home",
			"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN": token,
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
	port, _ := container.MappedPort(ctx, "8086/tcp")

	// Given
	influxUrl := fmt.Sprintf("http://%s:%d", host, port.Int())
	sender := selectBackend(map[string]string{
		Backend:           "influx",
		InfluxUrl:         influxUrl,
		InfluxToken:       token,
		InfluxOrg:         "samler",
		InfluxBucket:      "home",
		InfluxMeasurement: "power",
	})

	m := Measurement{
		Time:   time.Now(),
		Value:  42.5,
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

	client := influxdb2.NewClient(influxUrl, token)
	defer client.Close()
	query := client.QueryAPI("samler")

	result, err := query.Query(context.Background(),
		`import "date"
		 from(bucket: "home")
		   |> range(start: date.sub(d: 1m, from: now()), stop: now())
		   |> filter(fn: (r) => r["_measurement"] == "power")`)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Next() {
		t.Fatal()
	}

	resultValue := result.Record().Value()
	if resultValue != 42.5 {
		t.Fatal()
	}
}
