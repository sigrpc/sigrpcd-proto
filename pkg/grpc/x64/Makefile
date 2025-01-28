CC	= protoc
GO_OUT	= .
PROTO_PATH	= .
GO_GRPC_OUT	= .
SRCS	= message.proto

all: $(SRCS)
	$(CC) --proto_path=$(PROTO_PATH) --go_out=$(GO_OUT) --go_opt=paths=source_relative --go-grpc_out=$(GO_GRPC_OUT) --go-grpc_opt=paths=source_relative $(SRCS)

clean:
	$(RM) *.pb.go
