# generate proto buffer codes

GOBIN = ./build/
GO ?= latest
GORUN = env GO111MODULE=on go run
GOPROXY="https://goproxy.io,direct"

gen:
	@echo "Compiling proto..."
	@protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative proto/*.proto
	@echo "Generate proto buffer successfully"
# compile server and client
build:
	@echo "building server and client ..."
	@go build -o ./build/notary-server ./notary-server/.
	@go build -o ./build/notary-cli ./cli/.
	@echo "Finished compiling"

# install necessary binaries
install:
	@./scripts/download-binaries.sh

# install geth and start with a specified account
start:
	@./scripts/start-node.sh

stop:
	@./scripts/stop-node.sh

clean:
	@rm ~/.ethereum/ -rf


