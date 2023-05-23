.PHONY: proto
proto:
	protoc -I=${GOPATH}/src:. --gogo_out=paths=source_relative:. weberror/*.proto
	protoc -I=${GOPATH}/src:. --gogo_out=paths=source_relative:. webhttp/*.proto
