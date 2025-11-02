curl-produce: 
	curl -i -X POST -d '{"record": {"value": "5555"}}' http://localhost:8080/

curl-consume: 
	curl -i -X GET -d '{"offset": 0}' http://localhost:8080/



compile:
	protoc api/v1/*.proto \
	--go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	--proto_path=.
	
test:	
	go test -race ./...

tidy:
	go mod tidy
