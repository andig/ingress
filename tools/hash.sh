#!/bin/sh

HASH=$(git rev-parse HEAD | head -c8)
TAG=$(git describe --abbrev=0 --tags)

cat << EOF > cmd/ingress/version.go
package main

const (
	tag = "$TAG"
	hash = "$HASH"
)
EOF