#The .PHONY special target in a makefile is used to explicitly declare that a 
#target name is not a file name, but rather just a name for a sequence of commands 
#to be executed. This ensures the associated commands run every time the target is 
#requested, regardless of whether a file with that name exists in the directory. 

CONFIG_PATH = ${HOME}/.proglog/
TEST_FUNC_NAME = TestPickerNoSubConnAvailable
TEST_FILENAME = picker_test.go
TEST_FOLDER_NAME = loadbalance
TAG ?= 0.0.1

.PHONY: run	
run:
	go run cmd/main.go

.PHONY: build
build:
	go build -o main cmd/main.go 

.PHONY: test	
test:	
	go test -v ./...  -debug=true 

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...

.PHONY: view_coverage
view_coverage:
	go tool cover -html=coverage.out

.PHONY: test-single-func
test-single-func:	
	go test -v -run ^${FUNC_NAME} internal/${TEST_FOLDER_NAME}/${TEST_FILENAME}
 

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: debug
debug:
	dlv debug cmd/main.go

.PHONY: compile
compile:
	protoc api/v1/*.proto \
	--go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	--proto_path=.

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: gencert
gencert:
	cfssl gencert \
		-initca test/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server \
		test/server-csr.json | cfssljson -bare server

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		test/client-csr.json | cfssljson -bare client


	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		-cn="root" \
		test/client-csr.json | cfssljson -bare root-client

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		-cn="nobody" \
		test/client-csr.json | cfssljson -bare nobody-client

	mv *.pem *.csr ${CONFIG_PATH}

.PHONY: casbine-model-policy
 casbine-model-policy:
	cp test/model.conf $(CONFIG_PATH)/model.conf
	cp test/policy.csv $(CONFIG_PATH)/policy.csv

.PHONY: curl-produce
curl-produce: 
	curl -i -X POST -d '{"record": {"value": "5555"}}' http://localhost:8080/

.PHONY: curl-consume
curl-consume: 
	curl -i -X GET -d '{"offset": 2}' http://localhost:8080/

.PHONY: grpc-consume
grpc-consume:
	grpcurl -d '{"offset":2}' -plaintext localhost:4000 log.v1.Log/Consume



.PHONY: build-docker
build-docker:
	docker build -t github.com/nico-phil/proglog:$(TAG) .