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

package x64

import (
	"context"
	"io"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/msg"
	grpcclient "github.com/sigrpc/sigrpcd/pkg/domain/repository/grpc"
	"github.com/sigrpc/sigrpcd/pkg/grpc/x64"
	"google.golang.org/grpc"
)

type X64GRPCClient struct {
	Ctx          context.Context
	Client       x64.SigRPCClient
	ClientID     string
	StreamClient *x64.SigRPC_InvokeFuncClient
	isStreaming  bool
}

func NewClient(cc grpc.ClientConnInterface, ctx context.Context) grpcclient.GRPCClient {
	client := x64.NewSigRPCClient(cc)
	return &X64GRPCClient{
		Ctx:          ctx,
		Client:       client,
		StreamClient: nil,
		isStreaming:  false,
	}
}

func (c *X64GRPCClient) IsStreaming() bool {
	return c.isStreaming
}

func (c *X64GRPCClient) LoadLib(req *msg.LoadLibMsg) (*msg.LoadLibMsg, error) {
	resp, err := c.Client.LoadLib(c.Ctx, req.X64)
	if err != nil {
		return nil, err
	}
	loadlib := msg.LoadLibMsg{
		X64: resp,
	}
	return &loadlib, nil
}

func (c *X64GRPCClient) InvokeFunc(req *msg.InvokeFuncMsg) (*msg.InvokeFuncMsg, error) {
	if c.StreamClient == nil {
		stream, err := c.Client.InvokeFunc(c.Ctx)
		if err != nil {
			return nil, err
		}
		c.StreamClient = &stream
	}
	stream := *c.StreamClient
	err := stream.Send(req.X64)
	if err != nil {
		return nil, err
	}
	resp, err := stream.Recv()
	if err == nil {
		c.isStreaming = true
	} else if err == io.EOF {
		c.isStreaming = false
	}
	return &msg.InvokeFuncMsg{
		X64: resp,
	}, err
}

func (c *X64GRPCClient) PullPage(req *msg.PullPageMsg) (*msg.PullPageMsg, error) {
	resp, err := c.Client.PullPage(c.Ctx, req.X64)
	if err != nil {
		return nil, err
	}
	page := msg.PullPageMsg{
		X64: resp,
	}
	return &page, nil
}

func (c *X64GRPCClient) GetRPCType(header *msg.RPCHeader) uint32 {
	return header.X64.MsgType
}
