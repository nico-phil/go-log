curl-produce: 
	curl -i -X POST -d '{"record": {"value": "5555"}}' http://localhost:8080/

curl-consume: 
	curl -i -X GET -d '{"offset": 0}' http://localhost:8080/



compile: protoc api/v1/*.proto \ 
	--go_out=. \ 
	--go_opt=paths=golang \ 
	--proto_path=.
	
test:	
	go test -race ./...
