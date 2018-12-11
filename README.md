# ingress

`ingress` is a universal data ingestion and mapping daemon aimed at use with, but not limited to, the [Volkszähler] smart meter framework.

## Table of Contents

1. [Introduction](#Introduction)
1. [Quickstart](#Quickstart)
1. [Architecture](#Architecture)
2. [Data Sources](#Data%20Sources)
3. [Backlog](#Backlog)
4. [References](#References)

## Introduction

The `ingress` design goals are:
- ease of use: simplicity of configuration and auto-configuration
- flexibility: reusable functionality independent from data sources or targets
- extensibility: modular design
- performance: based on Go as compiled language (to be proven)

**Warning**: `ingress` is experimental. It may not work as described or I may lose interest in developing it any further. You're using it totally at your own risk.

## Quickstart

The following example describes how to connect [GoSDM] with [Volkszähler] for for data logging purposes.

### Prerequisites

- MQTT server like mosquitto. For testing purposes one of the available [public MQTT servers](https://github.com/mqtt/mqtt.github.io/wiki/public_brokers) can be used.
- [GoSDM] installed with Homie protocol enabled and connected to MQTT
- `ingress` installed

### Overview

This use case requires to connect the following components:

    +-------+      +-----------+      +---------+      +-------------+
    | GoSDM | ---> | Mosquitto | ---> | Ingress | ---> | Volkszähler |
    +-------+      +-----------+      +---------+      +-------------+

### Configuration

For this scenario `ingress` needs to connect mosquitto to [Volkszähler]. 

Since [GoSDM] is able to publish ModBus readings to MQTT, it is easiest to connect MQTT as input data *source* to an outbound *target* using HTTP as protocol for [Volkszähler]:

```yaml
sources:
- name: gosdm
  type: homie
targets:
- name: vz
  type: http
wires:
- source: gosdm
  target: vz
```

The configuration defines the data source and target with `wires` connecting the source `gosdm` to target `vz`.

For the data source- an MQTT server speaking the [Homie] protocol supported by [GoSDM]- we'll need to add the server address (here assuming [mosquitto] running on localhost):

```yaml
- name: gosdm
  type: homie
  url: tcp://localhost:1883
```

Since we're using a simple HTTP configuration for the target `vz`, we'll need to further define which HTTP requests should be executed:

```yaml
- name: vz
  type: http
  url: https://demo.volkszaehler.org/middleware.php/data/%name%.json
  method: POST
  headers:
    Content-type: application/json
    Accept: application/json
  payload: >-
    [[%timestamp%,%value%]]
```

The HTTP target `url` and the POST `payload` data are built from the received data using the define templates. Templates can contain the following variables:

- `name`: name of the input data reading
- `value`: value of the input data reading as formatted string
- `timestamp`: timestamp when the input data was received

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

Data source are not neccessarily directly connected to physical devices. For example, connecting [GoSDM] to `ingress` works best using an MQTT server.

## Data Sources

### Sources

- `mqtt`: MQTT server
    - `url`: server url including schema and port (default: tcp://localhost:1883)
    - `topic`: topic (default: #)
- `homie`: MQTT server with connected [Homie] clients like [GoSDM] (as of v0.8)
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

## Backlog

- WebSocket data source
- [GoElster](https://github.com/andig/goelster) CANBus data source

[Volkszähler]: https://volkszaehler.org
[GoSDM]: https://github.com/gonium/gosdm630
[Homie]: https://homieiot.github.io
[Mosquitto]: https://mosquitto.org