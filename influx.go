/*
SaMLer - Smart Meter data colletor at the edge
Copyright (C) 2025  Florian Heubeck

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
	"log"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func InitializeInflux(
	influxUrl string,
	influxToken string,
	influxOrg string,
	influxBucket string,
	influxMeasurement string,
) func(measurement Measurement) bool {
	fmt.Printf("Init Influx for %s at %s\n", influxOrg, influxUrl)

	influxClient := influxdb2.NewClient(influxUrl, influxToken)
	writeAPI := influxClient.WriteAPIBlocking(influxOrg, influxBucket)

	sender := func(measurement Measurement) bool {
		tags := map[string]string{
			"ident":  measurement.Ident,
			"unit":   measurement.Unit,
			"prefix": measurement.Prefix,
			"suffix": measurement.Suffix,
		}
		fields := map[string]interface{}{
			"value": measurement.Value,
		}
		debug("Sending to influx", &measurement)
		point := write.NewPoint(influxMeasurement, tags, fields, measurement.Time)
		err := writeAPI.WritePoint(context.Background(), point)
		if err != nil {
			log.Printf("Failed sending to influx %s\n", err)
			return false
		}
		return true
	}

	return sender
}
