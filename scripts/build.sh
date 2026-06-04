#!/bin/bash
set -e

echo "Building omo-switch..."

go build -o omo-switch ./cmd/omo-switch

echo "Build complete: ./omo-switch"
