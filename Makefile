
generate-keys:
	@./scripts/tls.sh

build-server:
	@./scripts/build-server.sh

build-client:
	@./scripts/build-client.sh

proto:
	@./scripts/proto.sh
