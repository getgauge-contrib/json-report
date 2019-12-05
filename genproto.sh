#!/bin/sh

#Using protoc version 3.0.0
cd gauge-proto
PATH=$PATH:$GOPATH/bin protoc --go_out=plugins=grpc:../gauge_messages messages.proto spec.proto services.proto
