# generate proto buffer codes

GO ?= latest
GORUN = env GO111MODULE=on go run

.PHONY: build install start stop gen
gen:
	@echo "Compiling proto..."
	@protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative proto/*.proto
	@echo "Generate proto buffer successfully"
# compile server and client
build:
	@echo "building server and client ..."
	@go build -o ./build/notary-server ./notary-server/.
	@go build -o ./build/notary-cli ./cli/.
	@cp ./notary-server/fabric/business/impl/config.yaml $(HOME)/.notary-samples/
	@echo "Finished compiling"

# install necessary binaries
install:
	@echo "start installing geth and fabric binaries"
	@./scripts/download-binaries.sh
	@echo "finished install binaries"

# install geth and start with a specified account
start: install build clean
	@cd scripts && exec ./start-node.sh

clean:
	@cd scripts && exec ./stop-node.sh

demo:
	@echo "demo process started"
	@cd scripts && exec ./demo.sh

start-server: start
	@cd build && exec nohup ./notary-server &