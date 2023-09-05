#!/usr/bin/env bash

set -euxo pipefail

go build -o bin/kafka 05-kafka/main.go

maelstrom test -w kafka --bin bin/kafka --node-count 1 --concurrency 2n --time-limit 20 --rate 1000