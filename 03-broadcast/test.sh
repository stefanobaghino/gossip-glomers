#!/usr/bin/env bash

set -euxo pipefail

go build -o bin/broadcast 03-broadcast/main.go

maelstrom test -w broadcast --bin bin/broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100 --nemesis partition