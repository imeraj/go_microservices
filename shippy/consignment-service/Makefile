build:
	protoc -I. --go_out=plugins=micro:$(GOPATH)/src/github.com/imeraj/go_microservices/shippy/consignment-service \
		proto/consignment/consignment.proto
	go build	