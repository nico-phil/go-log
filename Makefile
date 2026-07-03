#The .PHONY special target in a makefile is used to explicitly declare that a 
#target name is not a file name, but rather just a name for a sequence of commands 
#to be executed. This ensures the associated commands run every time the target is 
#requested, regardless of whether a file with that name exists in the directory. 

CONFIG_PATH = ${HOME}/.proglog/
TEST_NAME = "TestStoreAppenRead"
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


.PHONY: ${TEST_NAME}
${TEST_NAME}:
	go test  -run ^${TEST_NAME} go-log/log -v

.PHONY: test-clean
test-clean:
	rm -r proglog
 

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
	grpcurl -d '{"offset": 0}' -plaintext localhost:8402 log.v1.Log/Consume

.PHONY: grpc-produce
grpc-produce:
	grpcurl -d '{"record": {"value": "aGVsbG8gd29ybGQ="}}' -plaintext localhost:8400 log.v1.Log/Produce


.PHONY: grpc-produce-stream
grpc-produce-stream:
	{ \
		echo '{"record":{"value":"Zm9v"}}'; \
		echo '{"record":{"value":"YmF6"}}'; \
		echo '{"record":{"value":"YmFy"}}'; \
	} | grpcurl -plaintext \
		-d @ \
		localhost:8400 \
		log.v1.Log/ProduceStream

.PHONY: grpc-consume-stream
grpc-consume-stream:
	grpcurl \
		-plaintext \
		-d '{"offset":0}' \
		localhost:8401 \
		log.v1.Log/ConsumeStream

.PHONY: get-servers
get-servers:
	go run cmd/getservers/main.go

.PHONY: docker-build
docker-build:
	docker build -t github.com/nico-phil/proglog:$(TAG) .

.PHONY: docker-run
docker-run:
	docker run github.com/nico-phil/proglog:$(TAG) .

.PHONY: load-image
load-image:
	kind load docker-image github.com/nico-phil/proglog:$(TAG)

.PHONY: helm-i
helm-i: 
	helm install proglog deploy/proglog 


.PHONY: helm-uni
helm-uni: 
	-helm uninstall proglog
	-kubectl delete pvc datadir-proglog-0
	-kubectl delete pvc datadir-proglog-1
	-kubectl delete pvc datadir-proglog-2


.PHONY: deploy-local
deploy-local: docker-build load-image helm-i

.PHONY: delete-pvcs
delete-pvcs: 
	kubectl delete pvc datadir-proglog-0
	kubectl delete pvc datadir-proglog-1
	kubectl delete pvc datadir-proglog-2

.PHONY: get-pvcs
get-pvcs:
	kubectl get pvc  

.PHONY: forward-port
forward-port:
	 kubectl port-forward pod/proglog-0 8400:8400

