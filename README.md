# SaMLer

The SaMLer is reading SML messages from a serial device and publishes it to InfluxDB.

## Dependencies

It uses [libsml](https://github.com/volkszaehler/libsml) for the low level device and protocol handling, and has filesystem caching using [diskqueue](https://github.com/nsqio/go-diskqueue) to overcome (temporary) network issues.

## Use

