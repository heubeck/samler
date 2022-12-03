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
	os.Setenv(CachePath, "CachePath")
	os.Setenv(InfluxUrl, "InfluxUrl")
	os.Setenv(InfluxToken, "InfluxToken")
	os.Setenv(InfluxOrg, "InfluxOrg")
	os.Setenv(InfluxBucket, "InfluxBucket")
	os.Setenv(InfluxMeasurement, "InfluxMeasurement")

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
}
