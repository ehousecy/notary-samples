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
	@echo "Installing server and client "
	@go install github.com/ehousecy/notary-samples/
	@go install github.com/ehousecy/notary-samples/cli
	@echo "server and client are successfully installed"
	@echo "Installing geth and fabric"
	@go get github.com/ethereum/go-ethereum/tree/master/cmd/geth
	@echo "Installing Fabric v2.2"
	@curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.1 1.4.9

# install geth and start with a specified account
start:
	@./start.sh


