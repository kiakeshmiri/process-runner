#!/bin/bash
set -e


protoc \
  --proto_path=api/proto "api/proto/prunner.proto" \
  "--go_out=api/protogen" --go_opt=paths=source_relative \
  --go-grpc_opt=require_unimplemented_servers=false \
  "--go-grpc_out=api/protogen" --go-grpc_opt=paths=source_relative