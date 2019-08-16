# ingress

[![Build status](https://travis-ci.org/andig/ingress.svg?branch=master)](https://travis-ci.org/andig/ingress)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=HWZTN5AU8LSUC)

`ingress` is a universal data ingestion and mapping daemon aimed at use with, but not limited to, the [Volkszähler] smart meter framework.

## Table of Contents

1. [Introduction](#Introduction)
2. [Quickstart](#Quickstart)
3. [Architecture](#Architecture)
4. [Data Sources](#Data%20Sources)
5. [Mappings](#Mappings)
6. [Backlog](#Backlog)
7. [Frequently asked questions](#Frequently%20asked%20questions)

## Introduction

The `ingress` design goals are:
- ease of use: simplicity of configuration and auto-configuration
- flexibility: reusable functionality independent from data sources or targets
- extensibility: modular design
- performance: based on Go as compiled language (to be proven)

**Warning**: `ingress` is experimental. It may not work as described or I may lose interest in developing it any further. You're using it totally at your own risk.

## Quickstart

The following example describes how to connect [MBMD] with [Volkszähler] for for data logging purposes.

### Prerequisites

- MQTT server like mosquitto. For testing purposes one of the available [public MQTT servers](https://github.com/mqtt/mqtt.github.io/wiki/public_brokers) can be used.
- [MBMD] installed with Homie protocol enabled and connected to MQTT
- `ingress` installed

### Overview

This use case requires to connect the following components:

    +--------+      +-----------+      +---------+      +-------------+
    |  MBMD  | ---> | Mosquitto | ---> | Ingress | ---> | Volkszähler |
    +--------+      +-----------+      +---------+      +-------------+

### Configuration

For this scenario `ingress` needs to connect mosquitto to [Volkszähler]. 

Since [MBMD] is able to publish ModBus readings to MQTT, it is easiest to connect MQTT as input data *source* to an outbound *target* using HTTP as protocol for [Volkszähler]:

```yaml
sources:
- name: mbmd
  type: homie
targets:
- name: vz
  type: http
wires:
- source: mbmd
  target: vz
```

The configuration defines the data source and target with `wires` connecting the source `mbmd` to target `vz`.

For the data source- an MQTT server speaking the [Homie] protocol supported by [MBMD]- we'll need to add the server address (here assuming [mosquitto] running on localhost):

```yaml
- name: mbmd
  type: homie
  url: tcp://localhost:1883
```

Since we're using a simple HTTP configuration for the target `vz`, we'll need to further define which HTTP requests should be executed:

```yaml
- name: vz
  type: http
  url: https://demo.volkszaehler.org/middleware.php/data/{name}.json
  method: POST
  headers:
    Content-type: application/json
    Accept: application/json
  payload: >-
    [[{timestamp},{value}]]
```

The HTTP target `url` and the POST `payload` data are built from the received data using the define templates. Templates can contain the following variables:

- `name`: name of the input data reading
- `value`: value of the input data reading as formatted string
- `timestamp`: timestamp when the input data was received

#### Patterns

Whenever a data source or target supports configurable payloads, the `{variable:format}` syntax can be used to customize the output. Possible formats are as understood by golang `fmt.Printf()`. Timestamp formatting additionally supports s/ms/us/ns as format or any valid format of golang `time.Format()`. The following defaults are used:

- name: name as string (equivalent `{name:%s}`)
- value: value as float with 3 significant digits (equivalent `{value:%.3f}`)
- timestamp: unix timestamp in milliseconds (equivalent `{timestamp:s}`)

### Testing

Now start `ingress` and validate the parsed configuration:

    ingress -c config.yml -d

    2018/12/11 22:06:15 wiring: wiring homie -> vz
    2018/12/11 22:06:15 connector: starting homie
    2018/12/11 22:06:15 homie: connected to tcp://localhost:1883
    2018/12/11 22:06:15 homie: subscribed to topic homie

Simulate a [homie] device using `mosquitto_pub`:

    mosquitto_pub -t 'homie/meter1/zaehlwerk1/$properties' -m energy -r
    mosquitto_pub -t 'homie/meter1/zaehlwerk1/energy/$datatype' -m float -r

    2018/12/11 22:28:00 homie: discovered homie/meter1/zaehlwerk1/energy

Start sending actual data:

    mosquitto_pub -t homie/meter1/zaehlwerk1/energy -m 3.14

    2018/12/11 22:28:13 homie: recv (homie/meter1/zaehlwerk1/energy=3.14)
    2018/12/11 22:28:13 connector: recv from homie (energy=3.140000)
    2018/12/11 22:28:13 mapper: routing homie -> vz
    2018/12/11 22:28:13 vz: send POST https://demo.volkszaehler.org/middleware.php/data/energy.json (energy=3.140000)
    2018/12/11 22:28:13 vz: send failed POST 400 https://demo.volkszaehler.org/middleware.php/data/energy.json

## Architecture

The `ingress` data mapper architecture consists of data *sources* being mapped to *targets* by configurable *wires*.

    sources     wires         data mapper         wires    targets

    source 1  ---+       +-------------------+     +----> target 1
                 +---->  |                   |  ---+
    source 2  -------->  |      ingress      |  --------> target 2
                 +---->  |                   |  ---+
    source n  ---+       +-------------------+     +----> target n

*Data* is produced by or aquired from input data *sources*. 
Using configured *wires*, data is forwarded to output *targets*. *Wires* define which input and output sources/targets are connected. 
*Mappings* define which rules are applied for the connection. Any number of input *sources* can be connected to any number of output *targets* while using multiple mappings.

Data source are not neccessarily directly connected to physical devices. For example, connecting [MBMD] to `ingress` works best using an MQTT server.

## Data Sources

### Sources

- `mqtt`: MQTT server
    - `url`: server url including schema and port (default: tcp://localhost:1883)
    - `topic`: topic (default: #)
- `homie`: MQTT server with connected [Homie] clients like [MBMD] (as of v0.8)
    - `url`: server url including schema and port (default: tcp://localhost:1883)
    - `topic`: homie root topic (default: homie)

### Targets

- `mqtt`: MQTT server
    - `url`: server url including schema and port (default: tcp://localhost:1883)
    - `topic`: topic template (default: ingress/%name%)
- `http`: HTTP server
    - `url`: server url including schema and port
    - `headers`: HTTP headers as key:value pairs (e.g. Content-type: application/json)
    - `method`: HTTP method (default: GET)
    - `payload`: payload template for POST requests

## Actions

Actions are optional in the the `ingress` configuration. An action performs an operation on the data or metadata. Operations include storing data for later use (aggregation), removing data from processing (filters) or changing metadata (mapping).

Reusable actions are defined using the actions configuration key:

```yaml
actions:
- name: homie-to-volkszaehler
  type: mapping
  entries:
    energy: 014648c0-197f-11e8-9f68-afd012b00a13 # rename homie property name to volkszaehler uuid
```

To use a defined action it must be assigned to the respective wire:

```yaml
wires:
- source: mbmd
  target: vz
  actions:
  - homie-to-volkszaehler
```

### Rules

The following rules are applied depending on how mappings are assigned to a wire:

Configuration | Description
------------- | -----------
no mapping assigned | Mapping is treated as *pass through*, that is *any* source data is forwarded to the target
one or more mappings assigned | All assigned mappings are processed in order or definition, starting with the first mapping.<br/> For each mapping, the list of mapping entries is processed in sequence.<br/> If a matching mapping entry is found where `from` matches the received entity's name, the entity name is updated to `to`. Matching is performed by lower-case comparison. No further mapping rules are evaluated.<br/> If no mapping entry matches, the source data is *discarded*, i.e. removed from the wire.

## Frequently asked questions

1. `ingress` doesn't work as expected

   `ingress` is work in progress. To verify that `ingress` even understands your configuration run 

       ingress --dump

   to show what configuration ingress understood. If it looks correct feel free to open an issue. Always attach your configuration to the issue.

## Backlog

- data aggregation
- daemon mode
- web ui
- API
- auto-setup Volkszaehler form ingress or vice versa
- web socket data source
- [GoElster] CANBus data source

[Volkszähler]: https://volkszaehler.org
[Homie]: https://homieiot.github.io
[Mosquitto]: https://mosquitto.org
[MBMD]: https://github.com/volkszaehler/mbmd
[GoElster]: https://github.com/andig/goelster
