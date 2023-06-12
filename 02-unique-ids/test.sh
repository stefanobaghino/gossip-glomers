#!/usr/bin/env bash

set -euxo pipefail

go build -o bin/unique-ids 02-unique-ids/main.go

maelstrom test -w unique-ids --bin bin/unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition