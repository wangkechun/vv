gen_proto:
	protoc --go_out=plugins=grpc:. proto/vv.proto

.PHONY: gen_proto
