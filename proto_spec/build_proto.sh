#!/usr/bin/env bash

cmd="protoc -I/usr/local/include -I. \
            -I$GOPATH/src \
            -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis  \
            --go_out=plugins=grpc:. "

service="log_post_service.proto"
eval $cmd$service