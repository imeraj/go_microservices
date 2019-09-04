build:
	protoc -I. --go_out=plugins=micro:$(GOPATH)/src/github.com/imeraj/go_microservices/shippy/user-service \
		proto/user/user.proto
	GOOS=linux GOARCH=amd64 go build
	docker build -t user-service .

run:	
	docker run -p 50053:50051 \
		-e MICRO_SERVER_ADDRESS=:50051 \
		user-service
