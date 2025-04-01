#!/bin/bash -v

GOOS=linux GOARCH=amd64 go build -o bootstrap server_start_stop.go
zip function.zip bootstrap

aws lambda update-function-code \
  --function-name a_polyglot_start_stop \
  --zip-file fileb://function.zip \
  --publish