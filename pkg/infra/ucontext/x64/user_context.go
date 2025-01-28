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
	"io"
	"unsafe"

	"encoding/binary"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/cpu"
	"github.com/sigrpc/sigrpcd/pkg/domain/model/ucontext"
	cpucodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/cpu"
	ucontextcodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/ucontext"
)

type UserContextCodec struct {
	cpucodec.CPU
}

func NewCodec(cpucodec cpucodec.CPU) ucontextcodec.UserContext {
	return &UserContextCodec{cpucodec}
}

func (h *UserContextCodec) Encode(ctx *ucontext.UserContext) []byte {
	byteCPU := h.CPU.Encode(&cpu.CPU{
		X64: ctx.CPU.X64,
	})
	byteStackBottom := make([]byte, unsafe.Sizeof(ctx.StackBottom))
	binary.LittleEndian.PutUint64(byteStackBottom, ctx.StackBottom)

	return append(byteCPU, byteStackBottom...)
}

func (h *UserContextCodec) Decode(reader io.Reader) *ucontext.UserContext {
	ctx := ucontext.UserContext{}
	ctx.CPU = h.CPU.Decode(reader)
	if binary.Read(reader, binary.LittleEndian, &ctx.StackBottom) != nil {
		return nil
	}
	return &ctx
}
