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
#ifndef SML_CGO_H_
#define SML_CGO_H_

struct DeviceConfig {
    const char *name;
    int baudRate;
    const char *mode;
};

struct SmlValue{
    const char *value;
    const char *unit;
    const char *prefix;
    const char *ident;
    const char *suffix;
};

struct SmlData{
    struct SmlValue value;
};

typedef void (*SmlEvent)(struct SmlData message);

typedef struct {
    SmlEvent event;
} Callbacks;

int listen_to_device(struct DeviceConfig config, Callbacks callbacks);

extern void onSmlMessage(struct SmlData);
void propagateEvent(struct SmlData message);

#endif
