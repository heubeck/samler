# SaMLer

The SaMLer is reading SML messages produced by a smart meter from a serial device and publishes it to InfluxDB.

## Dependencies

It uses [libsml](https://github.com/volkszaehler/libsml) for the low level device and protocol handling, and has filesystem caching using [diskqueue](https://github.com/nsqio/go-diskqueue) to overcome (temporary) network issues.
Credits to these projects!

## Use

The dependency `libuuid` has to be installed on the target system, on Debian based systems availabe from package `uuid-runtime`.

There are pre-build binaries available with every [release](https://github.com/heubeck/samler/releases).

SaMLer is configured using environment variables, just run it, to let it print its options:

```shell
> ./samler.amd64
SaMLer; configure via ENV:
SAMLER_INFLUX_TOKEN (default: )
SAMLER_INFLUX_ORG (default: )
SAMLER_INFLUX_MEASUREMENT (default: power)
SAMLER_CACHE_PATH (default: /home/heubeck/.samler)
SAMLER_DEVICE_BAUD_RATE (default: 9600)
SAMLER_DEVICE_MODE (default: 8-N-1)
SAMLER_DEBUG (default: false)
SAMLER_INFLUX_URL (default: )
SAMLER_INFLUX_BUCKET (default: home)
SAMLER_DEVICE (default: /dev/ttyUSB0)
```

A minimalistic run script using [Influx Cloud](https://cloud2.influxdata.com/) may look like:

```shell
#!/bin/bash

export SAMLER_INFLUX_URL=https://region.provider.cloud2.influxdata.com
export SAMLER_INFLUX_TOKEN=thisIsVerySecret==
export SAMLER_INFLUX_ORG=your.influx.registered@mail.address

/opt/samler.arm-v7
```

Using systemd a service description may look like:

```shell
> cat /etc/systemd/system/samler.service
[Unit]
Description=SaMLer SmartMeter Data collector
StartLimitIntervalSec=0
[Service]
Type=simple
Restart=always
RestartSec=10
User=root
ExecStart=/opt/run_samler.sh

[Install]
WantedBy=multi-user.target
```

## Known restrictions & TODO

* By now, there's only a single serial mode supported what's reflected in the configuration defaults:
  _Baud rate `9600` and mode `8-N-1` (8 data bits, 1 stop bit, none parity)_
  That needs to be generalized.
* Smart Meter data is properitary to electric power meter, as that's the only one I have running SaMLer right now.
  Please file [issues](https://github.com/heubeck/samler/issues) with devices you'd like to read.
* Timing values are hard coded and should made configurable on demand.
* My C and Go skills are only rudimentary, don't hesitate to point out improvements.
* The only supported backend is InfluxDB by now, but it's prepared to support more in the future, just file an [issues](https://github.com/heubeck/samler/issues).

## Contribution

Yes, please. Looking forward to your ideas.

## Sample installation

At my place, I'm using the SaMLer on a Banana Pi 2 Zero connected to an USB IR sensor attached to the electric smart meter, sending to a free Influx2 Cloud account.
The visualization is done using a free Grafana Cloud account using the InfluxDB datasource.
At night, my WLAN is disabled, SaMLer caches the measurements on its "disk" during this network outage and sends it once the connection is recovered in the morning

![](static/SaMLer_IR.jpg)

![](static/SaMLer_Pi.jpg)

![](static/SaMLer_Wifi.jpg)

![](static/SaMLer_Grafana.png)
