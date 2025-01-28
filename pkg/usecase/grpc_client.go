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

package usecase

import (
	"bytes"
	"errors"
	"log"
	"net"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/msg"
	grpcclient "github.com/sigrpc/sigrpcd/pkg/domain/repository/grpc"
)

type GRPCClient struct {
	grpcclient.GRPCClient
	*MsgCodec
}

func NewGRPCClient(client grpcclient.GRPCClient, msgCodec *MsgCodec) *GRPCClient {
	return &GRPCClient{client, msgCodec}
}

func (c *GRPCClient) LoadLib(loadlib *msg.LoadLibMsg) (*msg.LoadLibMsg, error) {
	return c.GRPCClient.LoadLib(loadlib)
}

func (c *GRPCClient) InvokeFunc(invokeFunc *msg.InvokeFuncMsg) (*msg.InvokeFuncMsg, error) {
	return c.GRPCClient.InvokeFunc(invokeFunc)
}

func (c *GRPCClient) PullPage(page *msg.PullPageMsg) (*msg.PullPageMsg, error) {
	return c.GRPCClient.PullPage(page)
}

func (c *GRPCClient) InvokeRPC(conn net.Conn) ([]byte, error) {
	header, err := c.RPCHeaderCodec.Decode(conn)
	if err != nil {
		return nil, err
	}
	payload := make([]byte, header.X64.PayloadSize)
	readTotal := uint64(0)
	for readTotal < header.X64.PayloadSize {
		size, err := conn.Read(payload[readTotal:])
		if err != nil && size == 0 {
			return nil, err
		}
		readTotal += uint64(size)
	}
	reader := bytes.NewReader(payload)
	switch c.GetRPCType(header) {
	case msg.LOADLIB:
		if c.IsStreaming() {
			return nil, errors.New("LOADLIB does not support streaming")
		}
		req, err := c.LoadLibCodec.Decode(reader, header)
		if err != nil {
			return nil, err
		}
		resp, err := c.LoadLib(req)
		if err != nil {
			return nil, err
		}
		return c.LoadLibCodec.Encode(resp), nil
	case msg.INVOKEFUNC:
		req, err := c.InvokeFuncCodec.Decode(reader, header)
		if err != nil {
			return nil, err
		}
		resp, err := c.InvokeFunc(req)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return c.InvokeFuncCodec.Encode(resp), err
	case msg.PULLPAGE:
		if c.IsStreaming() {
			return nil, errors.New("PULLPAGE does not support streaming")
		}
		req, err := c.PullPageCodec.Decode(reader, header)
		if err != nil {
			return nil, err
		}
		resp, err := c.PullPage(req)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return c.PullPageCodec.Encode(resp), nil
	}
	return nil, errors.New("unsupported message")
}
