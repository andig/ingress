# ingress

`ingress` is a universal data ingestion and mapping component aimed at use with the [volkszaehler.org](https://volkszaehler.org) smart meter framework.

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

The following example describes how to connect GoSDM for data logging with Volkszähler.

### Prerequisites

- MQTT server like mosquitto. For testing purposes one of the available public MQTT servers can be used.
- GoSDM installed with Homie protocol enabled and connected to MQTT
- `ingress` installed

### Overview

This use case requires to connect the following components:

    +-------+      +-----------+      +---------+      +-------------+
    | GoSDM | ---> | Mosquitto | ---> | Ingress | ---> | Volkszähler |
    +-------+      +-----------+      +---------+      +-------------+

### Ingress Configuration

For this scenario `ingress` needs to connect mosquitto to volkszähler. This requires an input data *source* (mosquitto) and an outbound *target* that are wired together:

```yaml
sources:
- name: gosdm
  type: homie
targets:
- name: vz
  type: volkszaehler
wires:
- source: gosdm
  target: vz
mappings: # leave empty
```

The configuration defines the data source and target with `wiring` connecting the source `gosdm` to target `vz`.

## Architecture

The `ingress` data mapper architecture consists of data *sources* being mapped to *targets* by configurable *wires*.

    sources     wires         data mapper         wires    targets

    source 1  ---+       +-------------------+     +----> target 1
                 +---->  |                   |  ---+
    source 2  -------->  |      ingress      |  --------> target 2
                 +---->  |                   |  ---+
    source n  ---+       +-------------------+     +----> target n

*Data* is produces by or aquired from input data *sources*. Using configured *wires*, data is forwarded to output *targets*. *Wirings* define which input and output sources/targets are connected. *Mappings* define which rules are applied for the connection. Any number of input *sources* can be connected to any number of output *targets*.

Data source are not neccessarily directly connected to physical devices. For example, connecting [GoSDM](https://github.com/gonium/gosdm630) to `ingress` works best using an MQTT server 

## Data Sources

### Sources

- Generic MQTT adapter
- [GoSDM](https://github.com/gonium/gosdm630) ModBus adapter as of v0.8 using [Homie](https://https://homieiot.github.io) protocol

### Targets

- Generic MQTT adapter
- Volkszähler adapter

## Backlog

Future functionality may include the following topics.

Data sources/targets:
- HTTP adapter
- WebSocket adapter
- [GoElster](https://github.com/andig/goelster) CANBus adapter

## References 

[GoSDM]: <foo> bar [GoSDM](https://github.com/gonium/gosdm630)
