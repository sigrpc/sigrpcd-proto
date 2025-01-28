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
	"github.com/google/uuid"
	x64cpu "github.com/sigrpc/sigrpcd/pkg/infra/cpu/x64"
	x64page "github.com/sigrpc/sigrpcd/pkg/infra/page/x64"
	x64uctx "github.com/sigrpc/sigrpcd/pkg/infra/ucontext/x64"
	"github.com/sigrpc/sigrpcd/pkg/usecase"
)

func NewX64MsgCodec() (*usecase.MsgCodec, error) {
	clientID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	msgCodec := usecase.MsgCodec{}
	rpcHeaderCodec := NewRPCHeaderCodec(clientID.String())
	cpuCodec := usecase.NewCPUCodec(x64cpu.NewCodec())
	pageCodec := x64page.NewPageCodec()
	uctxCodec := usecase.NewUserContextCodec(
		x64uctx.NewCodec(&cpuCodec),
	)
	loadLibCodec := usecase.NewLoadLibCodec(
		NewLoadLibCodec(rpcHeaderCodec),
	)
	invokeFuncCodec := usecase.NewInvokeFuncCodec(
		NewInvokeFuncCodec(
			&uctxCodec,
			pageCodec,
			rpcHeaderCodec,
		),
	)
	pullPageCodec := usecase.NewPullPageCodec(
		NewPullPageCodec(
			pageCodec,
			rpcHeaderCodec,
		),
	)
	msgCodec.RPCHeaderCodec = usecase.NewRPCHeaderCodec(rpcHeaderCodec)
	msgCodec.LoadLibCodec = loadLibCodec
	msgCodec.InvokeFuncCodec = invokeFuncCodec
	msgCodec.PullPageCodec = pullPageCodec
	return &msgCodec, nil
}
