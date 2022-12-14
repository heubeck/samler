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
	"log"
	"os"
	"testing"
	"time"
)

func tempDir() string {
	dir, err := os.MkdirTemp("", "samler_test")
	if err != nil {
		log.Fatal()
	}
	return dir
}

func TestSuccessfullSending(t *testing.T) {
	// Given
	measurement := Measurement{
		Ident:  "ident",
		Unit:   "unit",
		Prefix: "prefix",
		Suffix: "suffix",
		Value:  42.23,
		Time:   time.Now(),
	}
	var sent Measurement
	messages := make(chan Measurement)
	send := func(m Measurement) bool {
		sent = m
		return true
	}
	RunSamler(messages, send, tempDir())

	// When
	messages <- measurement

	// Then
	if sent.Value != 42.23 {
		t.Error()
	}
}

func TestFailedSending(t *testing.T) {
	// Given
	measurement := Measurement{
		Ident:  "ident",
		Unit:   "unit",
		Prefix: "prefix",
		Suffix: "suffix",
		Value:  23.5,
		Time:   time.Now(),
	}
	result := true
	var sent Measurement
	messages := make(chan Measurement)
	send := func(m Measurement) bool {
		if !result {
			sent = m
		}
		result = !result
		return result
	}
	RunSamler(messages, send, tempDir())

	// When
	messages <- measurement

	// Then
	if sent.Value != 0 {
		t.Error()
	}

	time.Sleep(32 * time.Second)

	if sent.Value != 23.5 {
		t.Error()
	}
}
