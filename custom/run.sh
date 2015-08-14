#!/bin/bash

set -e
go build
docker build -t custom .
docker run --rm -p 8080:8080 custom
