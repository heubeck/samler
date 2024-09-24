/*
SaMLer - Smart Meter data colletor at the edge
Copyright (C) 2024  Florian Heubeck

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

/*
#cgo CFLAGS: -Ilibsml/sml/include/ -g -std=c99 -Wall -Wextra -pedantic
#cgo LDFLAGS: libsml/sml/lib/libsml.a -lm -luuid
#include <stdlib.h>
#include "samler.h"
*/
import "C"
import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	diskqueue "github.com/nsqio/go-diskqueue"
)

var Version string
var notice = fmt.Sprintf(`SaMLer v%s  Copyright (C) 2024  Florian Heubeck
This program comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it under certain conditions.
`, Version)

var debugFlag bool
var _messages = make(chan Measurement, 5000)

//export onSmlMessage
func onSmlMessage(msg C.struct_SmlData) {
	value := msg.value

	// fmt.Printf("%s Ident: %s, Value: %s %s\n", time.Now().Format("2006.01.02 15:04:05"), C.GoString(value.ident), C.GoString(value.value), C.GoString(value.unit))

	if val, err := strconv.ParseFloat(C.GoString(value.value), 32); err == nil {
		measure := Measurement{
			Ident:  C.GoString(value.ident),
			Unit:   C.GoString(value.unit),
			Prefix: C.GoString(value.prefix),
			Suffix: C.GoString(value.suffix),
			Value:  val,
			Time:   time.Now(),
		}
		debug("Sending to channel", &measure)
		_messages <- measure
	}
}

func dqLog(lvl diskqueue.LogLevel, f string, args ...interface{}) {
	log.Printf(lvl.String()+": "+f, args...)
}

const (
	Device            = "SAMLER_DEVICE"
	DeviceBaudRate    = "SAMLER_DEVICE_BAUD_RATE"
	DeviceMode        = "SAMLER_DEVICE_MODE"
	Debug             = "SAMLER_DEBUG"
	CachePath         = "SAMLER_CACHE_PATH"
	Backend           = "SAMLER_BACKEND"
	InfluxUrl         = "SAMLER_INFLUX_URL"
	InfluxToken       = "SAMLER_INFLUX_TOKEN"
	InfluxOrg         = "SAMLER_INFLUX_ORG"
	InfluxBucket      = "SAMLER_INFLUX_BUCKET"
	InfluxMeasurement = "SAMLER_INFLUX_MEASUREMENT"
	MySqlDSN          = "SAMLER_MYSQL_DSN"
	MySqlTable        = "SAMLER_MYSQL_TABLE"
	IdentFilter       = "SAMLER_IDENT_FILTER"
)

const (
	Influx = "influx"
	MySql  = "mysql"
)

var configOptions = map[string][]string{
	Device:            {"/dev/ttyUSB0"},
	DeviceBaudRate:    {"9600"},
	DeviceMode:        {"8-N-1"},
	Debug:             {"false"},
	CachePath:         {getUserHome() + "/.samler"},
	Backend:           {Influx, MySql},
	InfluxUrl:         {"-"},
	InfluxToken:       {"-"},
	InfluxOrg:         {"-"},
	InfluxBucket:      {"home"},
	InfluxMeasurement: {"power"},
	MySqlDSN:          {"-"},
	MySqlTable:        {"home_power"},
	IdentFilter:       {"-"},
}

func getUserHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return home
}

func readConfig() map[string]string {
	config := make(map[string]string)
	for key, def := range configOptions {
		value, isSet := os.LookupEnv(key)
		if isSet {
		} else if len(def) == 1 {
			value = def[0]
		} else {
			printHelpAndExit("Please set all values without default depending on the chosen backend")
		}
		config[key] = value
	}
	return config
}

func printHelpAndExit(hint string) {
	fmt.Println("# Configuration options, set them as ENV:")
	for key, dev := range configOptions {
		if len(dev) == 1 {
			fmt.Printf("%s (default: %s)\n", key, dev[0])
		} else {
			fmt.Printf("%s (options: %s)\n", key, strings.Join(dev[:], ", "))
		}
	}
	fmt.Printf("\n%s\n", hint)
	os.Exit(0)
}

func main() {
	fmt.Println(notice)
	config := readConfig()

	if flag, err := strconv.ParseBool(config[Debug]); err == nil {
		debugFlag = flag
	} else {
		printHelpAndExit(fmt.Sprintf("Illegal debug flag value %s: %s\n", config[Debug], err))
	}

	sendToBackend := selectBackend(config)

	// device config
	name := C.CString(config[Device])
	defer C.free(unsafe.Pointer(name))

	mode := C.CString(config[DeviceMode])
	defer C.free(unsafe.Pointer(mode))

	rate, err := strconv.Atoi(config[DeviceBaudRate])
	if err != nil {
		printHelpAndExit(fmt.Sprintf("Illegal baud rate value %s: %s\n", config[DeviceBaudRate], err))
	}
	baudRate := C.int(rate)

	deviceConfig := C.struct_DeviceConfig{
		name:     name,
		baudRate: baudRate,
		mode:     mode,
	}

	// callback
	callbacks := C.Callbacks{}
	callbacks.event = C.SmlEvent(C.propagateEvent)

	fmt.Println("Start Samler")
	RunSamler(_messages, sendToBackend, config[CachePath], config[IdentFilter])

	for {
		fmt.Printf("Listen to %s\n", config[Device])
		exitCode := int(C.listen_to_device(deviceConfig, callbacks))
		fmt.Printf("libsml exit: %d\n", exitCode)
		if exitCode != 0 {
			os.Exit(abs(exitCode))
		}
	}
}

func selectBackend(config map[string]string) func(measurement Measurement) bool {
	backend := config[Backend]
	switch backend {
	case MySql:
		return InitializeMySQL(
			config[MySqlDSN],
			config[MySqlTable],
		)
	case Influx:
		return InitializeInflux(
			config[InfluxUrl],
			config[InfluxToken],
			config[InfluxOrg],
			config[InfluxBucket],
			config[InfluxMeasurement],
		)
	default:
		printHelpAndExit(fmt.Sprintf("Unknown backend '%s', please select from [%s, %s]\n", backend, Influx, MySql))
		return func(m Measurement) bool { return false }
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// to be used in tests
func buildCSmlData(
	value string,
	unit string,
	prefix string,
	ident string,
	suffix string,
) C.struct_SmlData {
	return C.struct_SmlData{
		value: C.struct_SmlValue{
			value:  C.CString(value),
			unit:   C.CString(unit),
			prefix: C.CString(prefix),
			ident:  C.CString(ident),
			suffix: C.CString(suffix),
		},
	}
}
