// Copyright 2025 Keita HAGIWARA. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcclient "github.com/sigrpc/sigrpcd/pkg/infra/grpc/x64"
	"github.com/sigrpc/sigrpcd/pkg/infra/msg/x64"
	"github.com/sigrpc/sigrpcd/pkg/usecase"
)

func run(sock net.Listener, cc *grpc.ClientConn) {
	codec, err := x64.NewX64MsgCodec()
	if err != nil {
		log.Println(err)
		return
	}
	for {
		conn, err := sock.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
			defer conn.Close()
			defer cancel()
			grpcClient := grpcclient.NewClient(cc, ctx)
			sigRPCClient := usecase.NewGRPCClient(grpcClient, codec)
		read_next:
			resp, err := sigRPCClient.InvokeRPC(conn)
			if err != nil && err != io.EOF {
				log.Println(err)
				return
			}
			_, err = conn.Write(resp)
			if err != nil {
				return
			}
			if sigRPCClient.IsStreaming() {
				goto read_next
			}
		}()
	}
}

func main() {
	clientAddr := os.Getenv("RPC_CLIENT_ADDR")
	if len(clientAddr) == 0 {
		log.Println("RPC_CLIENT_ADDR is empty")
		return
	}
	clientNetwork := os.Getenv("RPC_CLIENT_NETWORK")
	if len(clientNetwork) == 0 {
		log.Println("RPC_CLIENT_NETWORK is empty")
		clientNetwork = "unix"
	}
	if _, err := os.Stat(clientAddr); err == nil {
		if err := os.RemoveAll(clientAddr); err != nil {
			log.Println(err)
			return
		}
	}
	addr := os.Getenv("RPC_STUB_ADDR")
	if len(addr) == 0 {
		log.Println("RPC_STUB_ADDR is empty")
		return
	}
	sock, err := net.Listen(clientNetwork, clientAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer sock.Close()
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxRecvMsgSizeCallOption{MaxRecvMsgSize: 0x7ffffffff}),
		grpc.WithDefaultCallOptions(grpc.MaxSendMsgSizeCallOption{MaxSendMsgSize: 0x7fffffff}))
	if err != nil {
		log.Println(err)
		return
	}
	defer cc.Close()
	run(sock, cc)
	if err := os.RemoveAll(clientAddr); err != nil {
		log.Println(err)
		return
	}
}
