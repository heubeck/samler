# SaMLer - Smart Meter data colletor at the edge
# Copyright (C) 2022  Florian Heubeck
# 
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# 
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# 
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

all: samler

libsml:
	make -C libsml/sml
	make -C libsml/examples

samler: libsml
	go build -ldflags "-X main.Version=$$VERSION"
	go test

.PHONY: clean libsml
clean:
	rm -f *.o
	rm -f samler
	make -C libsml clean
