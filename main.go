package main

/*
#cgo CFLAGS: -Ilibsml/sml/include/ -g -std=c99 -Wall -Wextra -pedantic
#cgo LDFLAGS: libsml/sml/lib/libsml.a -lm -luuid
#include <stdlib.h>
#include "samler.h"
*/
import "C"
import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"unsafe"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	diskqueue "github.com/nsqio/go-diskqueue"
)

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
	InfluxUrl         = "SAMLER_INFLUX_URL"
	InfluxToken       = "SAMLER_INFLUX_TOKEN"
	InfluxOrg         = "SAMLER_INFLUX_ORG"
	InfluxBucket      = "SAMLER_INFLUX_BUCKET"
	InfluxMeasurement = "SAMLER_INFLUX_MEASUREMENT"
)

var configOptions = map[string]string{
	Device:            "/dev/ttyUSB0",
	DeviceBaudRate:    "9600",
	DeviceMode:        "8-N-1",
	Debug:             "false",
	CachePath:         getUserHome() + "/.samler",
	InfluxUrl:         "",
	InfluxToken:       "",
	InfluxOrg:         "",
	InfluxBucket:      "home",
	InfluxMeasurement: "power",
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
		} else if def != "" {
			value = def
		} else {
			printHelpAndExit()
		}
		config[key] = value
	}
	return config
}

func printHelpAndExit() {
	fmt.Println("SaMLer; configure via ENV:")
	for key, dev := range configOptions {
		fmt.Printf("%s (default: %s)\n", key, dev)
	}
	os.Exit(0)
}

func main() {
	config := readConfig()

	if flag, err := strconv.ParseBool(config[Debug]); err == nil {
		debugFlag = flag
	} else {
		log.Printf("Illegal debug flag value %s: %s\n", config[Debug], err)
		printHelpAndExit()
	}

	fmt.Printf("Init Influx for %s at %s\n", config[InfluxOrg], config[InfluxUrl])

	influxClient := influxdb2.NewClient(config[InfluxUrl], config[InfluxToken])
	writeAPI := influxClient.WriteAPIBlocking(config[InfluxOrg], config[InfluxBucket])

	sendToInflux := func(measurement Measurement) bool {
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
		point := write.NewPoint(config[InfluxMeasurement], tags, fields, measurement.Time)
		err := writeAPI.WritePoint(context.Background(), point)
		if err != nil {
			log.Printf("Failed sending to influx %s\n", err)
			return false
		}
		return true
	}

	// config
	name := C.CString(config[Device])
	defer C.free(unsafe.Pointer(name))

	mode := C.CString(config[DeviceMode])
	defer C.free(unsafe.Pointer(mode))

	rate, err := strconv.Atoi(config[DeviceBaudRate])
	if err != nil {
		log.Printf("Illegal baud rate value %s: %s\n", config[DeviceBaudRate], err)
		printHelpAndExit()
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
	RunSamler(_messages, sendToInflux, config[CachePath])

	for {
		fmt.Printf("Listen to %s\n", config[Device])
		exitCode := int(C.listen_to_device(deviceConfig, callbacks))
		fmt.Printf("libsml exit: %d\n", exitCode)
		if exitCode != 0 {
			os.Exit(abs(exitCode))
		}
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
