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
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/cpu"
	"github.com/sigrpc/sigrpcd/pkg/domain/model/msg"
	"github.com/sigrpc/sigrpcd/pkg/domain/model/page"
	"github.com/sigrpc/sigrpcd/pkg/domain/model/ucontext"
	msgcodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/msg"
	pagecodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/page"
	ucontextcodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/ucontext"
	"github.com/sigrpc/sigrpcd/pkg/grpc/x64"
)

type InvokeFuncCodec struct {
	ucontextcodec.UserContext
	pagecodec.Page
	msgcodec.RPCHeader
}

func NewInvokeFuncCodec(
	userContextCodec ucontextcodec.UserContext,
	pageCodec pagecodec.Page,
	rpcHeaderCodec msgcodec.RPCHeader) msgcodec.InvokeFunc {
	return &InvokeFuncCodec{
		UserContext: userContextCodec,
		Page:        pageCodec,
		RPCHeader:   rpcHeaderCodec,
	}
}

func (h *InvokeFuncCodec) Encode(invokeFunc *msg.InvokeFuncMsg) []byte {
	byteInvokeFuncID := make([]byte, unsafe.Sizeof(invokeFunc.X64.InvokefuncId))
	binary.LittleEndian.PutUint64(byteInvokeFuncID, invokeFunc.X64.InvokefuncId)
	bytePayload := byteInvokeFuncID
	x64CPU := cpu.CPU{
		X64: invokeFunc.X64.Ctx.Cpu,
	}
	ctx := ucontext.UserContext{
		CPU:         &x64CPU,
		StackBottom: invokeFunc.X64.Ctx.StackBottom,
	}
	byteUserContext := h.UserContext.Encode(&ctx)
	bytePayload = append(bytePayload, byteUserContext...)
	for _, x64page := range invokeFunc.X64.Page {
		page := page.Page{
			X64: x64page,
		}
		bytePage := h.Page.Encode(&page)
		bytePayload = append(bytePayload, bytePage...)
	}

	invokeFunc.X64.Header.PayloadSize = uint64(len(bytePayload))
	header := msg.RPCHeader{
		X64: invokeFunc.X64.Header,
	}
	byteHeader := h.RPCHeader.Encode(&header)
	byteInvokeFunc := byteHeader
	byteInvokeFunc = append(byteInvokeFunc, bytePayload...)

	return byteInvokeFunc
}

func (h *InvokeFuncCodec) Decode(reader io.Reader, header *msg.RPCHeader) (*msg.InvokeFuncMsg, error) {
	invokeFunc := msg.InvokeFuncMsg{
		X64: &x64.InvokeFuncMsg{
			Header:       header.X64,
			InvokefuncId: 0,
			Ctx:          &x64.UserContext{},
			Page:         nil,
		},
	}
	err := binary.Read(reader, binary.LittleEndian, &invokeFunc.X64.InvokefuncId)
	if err != nil {
		return nil, err
	}
	userContext := h.UserContext.Decode(reader)
	invokeFunc.X64.Ctx.Cpu = userContext.CPU.X64
	invokeFunc.X64.Ctx.StackBottom = userContext.StackBottom
	for {
		p := h.Page.Decode(reader)
		if p == nil {
			break
		}
		invokeFunc.X64.Page = append(invokeFunc.X64.Page, p.X64)
	}

	return &invokeFunc, nil
}
