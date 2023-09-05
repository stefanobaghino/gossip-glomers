#!/usr/bin/env bash

set -euxo pipefail

go build -o bin/counter 04-counter/main.go

maelstrom test -w g-counter --bin bin/counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition