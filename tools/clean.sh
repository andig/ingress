#!/bin/sh
mosquitto_sub -t "#" -v -W 1 | while read line _; do echo $line; mosquitto_pub -t "$line" -r -n; done
