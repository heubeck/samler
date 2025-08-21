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
	"os"
	"testing"
)

func TestAbs(t *testing.T) {
	if abs(-1) != 1 {
		t.Error()
	}

	if abs(0) != 0 {
		t.Error()
	}

	if abs(1) != 1 {
		t.Error()
	}
}

func TestOnSmlMessage(t *testing.T) {
	// Given
	smlData := buildCSmlData(
		"23.5", "Tt", "pfx", "667", "sfx",
	)

	// When
	onSmlMessage(smlData)

	// Then
	measurement := <-_messages
	if measurement.Value != 23.5 {
		t.Error()
	}
	if measurement.Unit != "Tt" {
		t.Error()
	}
	if measurement.Prefix != "pfx" {
		t.Error()
	}
	if measurement.Ident != "667" {
		t.Error()
	}
	if measurement.Suffix != "sfx" {
		t.Error()
	}
}

func TestReadOverriddenConfig(t *testing.T) {
	// Given
	os.Setenv(Device, "Device")
	os.Setenv(DeviceBaudRate, "DeviceBaudRate")
	os.Setenv(DeviceMode, "DeviceMode")
	os.Setenv(Debug, "Debug")
	os.Setenv(Backend, "Influx")
	os.Setenv(CachePath, "CachePath")
	os.Setenv(InfluxUrl, "InfluxUrl")
	os.Setenv(InfluxToken, "InfluxToken")
	os.Setenv(InfluxOrg, "InfluxOrg")
	os.Setenv(InfluxBucket, "InfluxBucket")
	os.Setenv(InfluxMeasurement, "InfluxMeasurement")
	os.Setenv(MySqlDSN, "MySqlDSN")
	os.Setenv(MySqlTable, "MySqlTable")
	os.Setenv(IdentFilter, "IdentFilter")

	// When
	config := readConfig()

	// Then
	if config[Device] != "Device" {
		t.Fatal()
	}
	if config[DeviceBaudRate] != "DeviceBaudRate" {
		t.Fatal()
	}
	if config[DeviceMode] != "DeviceMode" {
		t.Fatal()
	}
	if config[Debug] != "Debug" {
		t.Fatal()
	}
	if config[Backend] != "Influx" {
		t.Fatal()
	}
	if config[CachePath] != "CachePath" {
		t.Fatal()
	}
	if config[InfluxUrl] != "InfluxUrl" {
		t.Fatal()
	}
	if config[InfluxToken] != "InfluxToken" {
		t.Fatal()
	}
	if config[InfluxOrg] != "InfluxOrg" {
		t.Fatal()
	}
	if config[InfluxBucket] != "InfluxBucket" {
		t.Fatal()
	}
	if config[InfluxMeasurement] != "InfluxMeasurement" {
		t.Fatal()
	}
	if config[MySqlDSN] != "MySqlDSN" {
		t.Fatal()
	}
	if config[MySqlTable] != "MySqlTable" {
		t.Fatal()
	}
	if config[IdentFilter] != "IdentFilter" {
		t.Fatal()
	}
}
