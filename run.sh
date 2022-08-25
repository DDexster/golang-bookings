#!/bin/zsh

go build -o bookings cmd/web/*.go && ./bookings -dbuser=ddexster -dbname=bookings -production=false -cache=false