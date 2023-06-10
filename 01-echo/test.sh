#!/usr/bin/env bash

set -euxo pipefail

go build -o bin/echo 01-echo/main.go

/opt/maelstrom/maelstrom test -w echo --bin bin/echo --node-count 1 --time-limit 10