#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o bootstrap server_start_stop.go
zip function.zip bootstrap
