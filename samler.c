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
#include <stdio.h>
#include <fcntl.h>
#include <unistd.h>
#include <getopt.h>
#include <ctype.h>
#include <errno.h>
#include <termios.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <math.h>
#include <sys/ioctl.h>

#include "libsml/sml/include/sml/sml_file.h"
#include "libsml/sml/include/sml/sml_transport.h"
#include "libsml/sml/include/sml/sml_value.h"

#include "libsml/examples/unit.h"
#include "samler.h"

/*
* Most of the following is taken from the libsml examples, many thanks to the volkszaehler project.
*/

// Registered callback
SmlEvent event;

int serial_port_open(struct DeviceConfig *deviceConfig) {
	int bits;
	struct termios config;
	memset(&config, 0, sizeof(config));

#ifdef O_NONBLOCK
	int fd = open(deviceConfig->name, O_RDWR | O_NOCTTY | O_NONBLOCK);
#else
	int fd = open(deviceConfig->name, O_RDWR | O_NOCTTY | O_NDELAY);
#endif
	if (fd < 0) {
		fprintf(stderr, "error: open(%s): %s\n", deviceConfig->name, strerror(errno));
		return -1;
	}

	// set RTS
	ioctl(fd, TIOCMGET, &bits);
	bits |= TIOCM_RTS;
	ioctl(fd, TIOCMSET, &bits);

	tcgetattr(fd, &config);

	if (strcmp(deviceConfig->mode, "8-N-1") == 0) {
		// set 8-N-1
		config.c_iflag &= ~(IGNBRK | BRKINT | PARMRK | ISTRIP | INLCR | IGNCR
				| ICRNL | IXON);
		config.c_oflag &= ~OPOST;
		config.c_lflag &= ~(ECHO | ECHONL | ICANON | ISIG | IEXTEN);
		config.c_cflag &= ~(CSIZE | PARENB | PARODD | CSTOPB);
		config.c_cflag |= CS8;
	} else {
		fprintf(stderr, "error: unsupported mode: %s\n", deviceConfig->mode);
		return -1;
	}

	if (deviceConfig->baudRate == 9600) {
		// set speed to 9600 baud
		cfsetispeed(&config, B9600);
		cfsetospeed(&config, B9600);
	} else {
		fprintf(stderr, "error: unsupported baud rate: %d\n", deviceConfig->baudRate);
		return -1;
	}

	tcsetattr(fd, TCSANOW, &config);
	return fd;
}

void transport_receiver(unsigned char *buffer, size_t buffer_len) {
	int i;
	// the buffer contains the whole message, with transport escape sequences.
	// these escape sequences are stripped here.
	sml_file *file = sml_file_parse(buffer + 8, buffer_len - 16);

	for (i = 0; i < file->messages_len; i++) {
		sml_message *message = file->messages[i];

		if (*message->message_body->tag == SML_MESSAGE_GET_LIST_RESPONSE) {
			sml_list *entry;
			sml_get_list_response *body;
			body = (sml_get_list_response *) message->message_body->data;
			for (entry = body->val_list; entry != NULL; entry = entry->next) {

				char identString[10] = "";
				char prefixString[5] = "";
				char suffixString[5] = "";
				char valueString[20] = "";
				char unitString[5] = "";

				if (!entry->value) { // do not crash on null value
					fprintf(stderr, "Error in data stream. entry->value should not be NULL. Skipping this.\n");
					continue;
				}
				snprintf(prefixString, 5, "%d-%d",
						entry->obj_name->str[0], entry->obj_name->str[1]
				);
				snprintf(identString, 10, "%d.%d.%d",
						entry->obj_name->str[2], entry->obj_name->str[3], entry->obj_name->str[4]
				);
				snprintf(suffixString, 5, "%d", entry->obj_name->str[5]);

				if (entry->value->type == SML_TYPE_OCTET_STRING) {

					char *str;
					snprintf(valueString, 20, "%s", sml_value_to_strhex(entry->value, &str, true));
					free(str);

				} else if (entry->value->type == SML_TYPE_BOOLEAN) {

					snprintf(valueString, 20, "%s", entry->value->data.boolean ? "true" : "false");

				} else if (((entry->value->type & SML_TYPE_FIELD) == SML_TYPE_INTEGER) ||
						((entry->value->type & SML_TYPE_FIELD) == SML_TYPE_UNSIGNED)) {

					double value = sml_value_to_double(entry->value);
					int scaler = (entry->scaler) ? *entry->scaler : 0;
					int prec = -scaler;
					if (prec < 0)
						prec = 0;
					value = value * pow(10, scaler);

					const char *unit = NULL;
					if (entry->unit) { // do not crash on null (unit is optional)
						unit = dlms_get_unit((unsigned char) *entry->unit);
					}

					snprintf(valueString, 20, "%.*f", prec, value);
					snprintf(unitString, 5,  "%s", unit != NULL ? unit : "");
				}

				struct SmlValue smlValue;
				memset(&smlValue, 0, sizeof(smlValue));
				smlValue.ident = identString;
				smlValue.value = valueString;
				smlValue.unit = unitString;
				smlValue.prefix = prefixString;
				smlValue.suffix = suffixString;

				struct SmlData smlData;
				memset(&smlData, 0, sizeof(smlData));
				smlData.value = smlValue;
				event(smlData);
			}
		}
	}

	// free the malloc'd memory
	sml_file_free(file);
}

int listen_to_device(struct DeviceConfig config, Callbacks callbacks){
	event = callbacks.event;
	// open serial port
	int fd = serial_port_open(&config);
	if (fd < 0) {
		// error message is printed by serial_port_open()
		return fd;
	}

	// listen on the serial device, this call is blocking.
	sml_transport_listen(fd, &transport_receiver);
	close(fd);

	return 0;
}

// handler function, passed from the Go part as callback
void propagateEvent(struct SmlData msg) {
	onSmlMessage(msg);
}
