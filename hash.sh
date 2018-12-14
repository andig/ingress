#!/bin/bash

HASH=$(git rev-parse HEAD | head -c8)
TAG=$(git describe --abbrev=0 --tags)

cat << EOF > cmd/ingress/version.go
package main

const (
	tag = "$TAG"
	version = "$HASH (https://github.com/andig/ingress/commit/$HASH)"
)
EOF