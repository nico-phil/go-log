#The .PHONY special target in a makefile is used to explicitly declare that a 
#target name is not a file name, but rather just a name for a sequence of commands 
#to be executed. This ensures the associated commands run every time the target is 
#requested, regardless of whether a file with that name exists in the directory. 

.PHONY: run	
run:
	go run cmd/main.go
.PHONY: test	
test:	
	go test -race ./...

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

.PHONY: curl-produce
curl-produce: 
	curl -i -X POST -d '{"record": {"value": "5555"}}' http://localhost:8080/

.PHONY: curl-consume
curl-consume: 
	curl -i -X GET -d '{"offset": 2}' http://localhost:8080/

.PHONY: grpc-consume
grpc-consume:
	grpcurl -d '{"offset":2}' -plaintext localhost:4000 log.v1.Log/Consume





