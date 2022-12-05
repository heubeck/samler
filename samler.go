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
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	diskqueue "github.com/nsqio/go-diskqueue"
)

type Measurement struct {
	Ident  string
	Unit   string
	Prefix string
	Suffix string
	Value  float64
	Time   time.Time
}

type samler struct {
	messageChannel chan Measurement
	send           func(Measurement) bool
	cacheLocation  string
}

var memo = make(map[string]Measurement)

func debug(msg string, measure *Measurement) {
	if debugFlag {
		fmt.Printf("%s: %s - %f\n", msg, measure.Ident, measure.Value)
	}
}

func RunSamler(
	messageChannel chan Measurement,
	send func(Measurement) bool,
	cacheLocation string,
) {
	samler := samler{
		messageChannel: messageChannel,
		send:           send,
		cacheLocation:  cacheLocation,
	}
	go processLoop(&samler)
}

func shouldSendAndMemorize(measure Measurement) bool {
	key := fmt.Sprintf("%s#%s#%s", measure.Prefix, measure.Ident, measure.Suffix)
	previous, ok := memo[key]
	if !ok || previous.Value != measure.Value || previous.Time.Add(60*time.Second).Before(time.Now()) {
		debug("Memorized", &measure)
		memo[key] = measure
		return true
	} else {
		debug("Skipped", &measure)
		return false
	}
}

func processLoop(ctx *samler) {
	fmt.Printf("Init DiskQueue at %s", ctx.cacheLocation)
	if err := os.MkdirAll(ctx.cacheLocation, fs.ModePerm); err != nil {
		log.Fatal(err)
	}
	diskQueue := diskqueue.New("cached", ctx.cacheLocation, 10485760, 4, 1<<10, 4096, 10*time.Second, dqLog)
	defer diskQueue.Close()

	readChan := diskQueue.ReadChan()
	peekChan := diskQueue.PeekChan()
	circuitOpen := false

	send := func(measure Measurement) bool {
		success := ctx.send(measure)
		if !success {
			circuitOpen = true
			log.Printf("Ciruit open")
			go func() {
				time.Sleep(30 * time.Second)
				circuitOpen = false
				log.Printf("Ciruit closed")
			}()
		}
		return success
	}

	writeToDisk := func(measure Measurement) {
		debug("Writing to disk", &measure)
		if jsonData, err := json.Marshal(measure); err == nil {
			diskQueue.Put(jsonData)
		} else {
			log.Fatal("Could not serialize measure", err)
		}
	}

	peekFromDisk := func() Measurement {
		message := <-peekChan
		var measure Measurement
		if err := json.Unmarshal(message, &measure); err != nil {
			log.Fatal("Failed to deserialize msg", err)
		}
		debug("Read from disk", &measure)
		return measure
	}

	removeLastPeeked := func() {
		<-readChan
	}

	// process disk messages
	go func() {
		for {
			if circuitOpen {
				time.Sleep(30 * time.Second)
			}
			measurement := peekFromDisk()
			if send(measurement) {
				removeLastPeeked()
			}
		}
	}()

	for {
		measurement := <-ctx.messageChannel

		if shouldSendAndMemorize(measurement) {
			if circuitOpen || !send(measurement) {
				writeToDisk(measurement)
			}
		}
	}
}
